from __future__ import annotations

import importlib.util
import io
import json
import sys
import unittest
import urllib.error
import urllib.parse
from pathlib import Path
from unittest import mock

MODULE_PATH = Path(__file__).parents[1] / "slack_api_delta.py"
SPEC = importlib.util.spec_from_file_location("slack_api_delta", MODULE_PATH)
assert SPEC and SPEC.loader
slack = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = slack
SPEC.loader.exec_module(slack)

XOXC = "xoxc-SYNTHETIC_DO_NOT_DISCLOSE_123"
XOXD = "xoxd-SYNTHETIC_DO_NOT_DISCLOSE_456"
CREDS = slack.Credentials(XOXC, XOXD)
START = slack.parse_iso("2026-08-08T10:00:00Z")
END = slack.parse_iso("2026-08-08T11:00:00Z")


class FakeResponse:
    def __init__(self, payload):
        self.data = json.dumps(payload).encode()

    def __enter__(self):
        return self

    def __exit__(self, *args):
        return False

    def read(self, limit):
        return self.data


class QueueOpener:
    def __init__(self, payloads):
        self.payloads = list(payloads)
        self.requests = []

    def __call__(self, request, **kwargs):
        self.requests.append(request)
        item = self.payloads.pop(0)
        if isinstance(item, Exception):
            raise item
        return FakeResponse(item)


def search_payload(matches, cursor="", page=None, pages=None):
    messages = {"matches": matches}
    if page is not None:
        messages["paging"] = {"page": page, "pages": pages}
    return {"ok": True, "messages": messages, "response_metadata": {"next_cursor": cursor}}


def match(ts, text="hello", channel="C12345678", user="U08BDJAMSRZ", thread_ts=None):
    result = {
        "ts": ts,
        "text": text,
        "channel": {"id": channel, "name": "private-name-must-not-leak"},
        "user_id": user,
        "permalink": f"https://example.slack.com/archives/{channel}/p{ts.replace('.', '')}",
    }
    if thread_ts:
        result["thread_ts"] = thread_ts
    return result


class CollectorTests(unittest.TestCase):
    @staticmethod
    def http_error(status=429, retry_after="1", reason="rate limited", body=b""):
        headers = {"Retry-After": retry_after} if retry_after is not None else {}
        return urllib.error.HTTPError(
            "https://slack.com/api/search.messages", status, reason, headers, io.BytesIO(body)
        )

    def test_429_retries_with_retry_after_then_succeeds_without_real_sleep(self):
        sleeps = []
        opener = QueueOpener([
            self.http_error(retry_after="2"),
            {"ok": True},
        ])
        client = slack.SlackClient(CREDS, opener=opener, sleeper=sleeps.append)

        self.assertEqual(client.get("auth.test"), {"ok": True})
        self.assertEqual(sleeps, [2.0])
        self.assertEqual(len(opener.requests), 2)

    def test_429_retry_exhaustion_fails_closed_with_bounded_wait(self):
        sleeps = []
        opener = QueueOpener([
            self.http_error(retry_after="3")
            for _ in range(slack.MAX_HTTP_ATTEMPTS)
        ])
        result = slack.collect(
            slack.SlackClient(CREDS, opener=opener, sleeper=sleeps.append),
            CREDS, START, END, slack.DEFAULT_USER,
        )

        self.assertTrue(result["coverage"]["fail_closed"])
        self.assertEqual(result["coverage"]["errors"], ["http_error"])
        self.assertEqual(result["events"], [])
        self.assertEqual(len(opener.requests), slack.MAX_HTTP_ATTEMPTS)
        self.assertEqual(sum(sleeps), 3.0 * (slack.MAX_HTTP_ATTEMPTS - 1))
        self.assertLessEqual(sum(sleeps), slack.MAX_TOTAL_RETRY_WAIT_SECONDS)

    def test_429_error_does_not_disclose_secret_reason_headers_or_body(self):
        header_secret = "header-" + XOXC
        error = self.http_error(
            retry_after="1 " + header_secret,
            reason="rate limited " + XOXC,
            body=("private body " + XOXD).encode(),
        )
        result = slack.collect(
            slack.SlackClient(CREDS, QueueOpener([error]), sleeper=lambda _: None),
            CREDS, START, END, slack.DEFAULT_USER,
        )

        serialized = json.dumps(result)
        self.assertEqual(result["coverage"]["errors"], ["http_error"])
        self.assertTrue(result["coverage"]["fail_closed"])
        self.assertNotIn(XOXC, serialized)
        self.assertNotIn(XOXD, serialized)
        self.assertNotIn(header_secret, serialized)

    def test_search_views_use_with_modifier_for_involving_and_preserve_from(self):
        views = dict(slack.search_views(slack.DEFAULT_USER, "2026-08-07", "2026-08-09"))
        self.assertEqual(
            views["authored"],
            "from:<U08BDJAMSRZ> after:2026-08-07 before:2026-08-09",
        )
        self.assertEqual(
            views["involving"],
            "with:<@U08BDJAMSRZ> after:2026-08-07 before:2026-08-09",
        )
        self.assertNotEqual(views["involving"].split()[0], "<@U08BDJAMSRZ>")

        opener = QueueOpener([search_payload([]), search_payload([])])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["complete"])
        queries = [urllib.parse.parse_qs(urllib.parse.urlsplit(request.full_url).query)["query"][0]
                   for request in opener.requests]
        self.assertEqual(queries, [views["authored"], views["involving"]])

    def test_cursor_pagination_exact_filter_dedup_and_snippet_cap(self):
        long_secret_text = "I will review this. " + ("é" * 300) + " " + XOXC + " " + XOXD
        opener = QueueOpener([
            search_payload([
                match("1786183199.999", "too early"),
                match("1786185000.000", long_secret_text, user="U08BDJAMSRZ", thread_ts="1786184000.000"),
            ], cursor="NEXT"),
            search_payload([match("1786186200.000", "I'll handle the review", user="U08BDJAMSRZ")]),
            search_payload([
                match("1786185000.000", long_secret_text, user="U08BDJAMSRZ", thread_ts="1786184000.000"),
                match("1786186800.001", "too late"),
            ]),
        ])
        client = slack.SlackClient(CREDS, opener=opener)
        result = slack.collect(client, CREDS, START, END, slack.DEFAULT_USER)

        coverage = result["coverage"]
        self.assertTrue(coverage["complete"])
        self.assertFalse(coverage["incomplete"])
        self.assertEqual(coverage["pages"], 3)
        self.assertEqual(coverage["matches_scanned"], 5)
        self.assertEqual(coverage["exact_matches"], 3)
        self.assertEqual(coverage["candidate_matches"], 3)
        self.assertEqual(coverage["filtered_noise"], 0)
        self.assertEqual(coverage["deduplicated"], 1)
        self.assertEqual(coverage["returned"], 2)
        self.assertEqual(len(result["events"]), 2)
        first = result["events"][0]
        self.assertEqual(first["id"], "slack:C12345678:1786185000.000")
        self.assertEqual(first["thread_ts"], "1786184000.000")
        self.assertLessEqual(len(first["snippet"]), 240)
        self.assertIn("possible_commitment", first["classification_hints"])
        serialized = json.dumps(result)
        self.assertNotIn(XOXC, serialized)
        self.assertNotIn(XOXD, serialized)
        self.assertNotIn("private-name-must-not-leak", serialized)
        self.assertIn("cursor=NEXT", opener.requests[1].full_url)

    def test_legacy_page_pagination(self):
        opener = QueueOpener([
            search_payload([match("1786185000.000", "I will handle this")], page=1, pages=2),
            search_payload([match("1786185100.000", "Done")], page=2, pages=2),
            search_payload([], page=1, pages=1),
        ])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["complete"])
        self.assertEqual(result["coverage"]["pages"], 3)
        self.assertIn("page=2", opener.requests[1].full_url)

    def test_candidate_filter_runs_before_event_cap(self):
        noise = [match(f"1786185{i:03d}.000", "routine status FYI") for i in range(100)]
        relevant = [
            match("1786186100.000", "Je vais le faire"),
            match("1786186200.000", "Je m’en charge demain"),
        ]
        opener = QueueOpener([search_payload(noise + relevant), search_payload([])])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["complete"])
        self.assertEqual(result["coverage"]["exact_matches"], 102)
        self.assertEqual(result["coverage"]["candidate_matches"], 2)
        self.assertEqual(result["coverage"]["filtered_noise"], 100)
        self.assertEqual(result["coverage"]["returned"], 2)
        self.assertNotIn("event_cap_reached", result["coverage"]["errors"])

    def test_english_and_french_candidate_phrase_coverage(self):
        authored = [
            "I'll do it", "je vais vérifier", "je m'en occupe", "je m’en charge",
            "demain", "plus tard", "fixed", "résolu",
        ]
        inbound = [
            "can you check", "could you help", "would you review", "peux-tu valider",
            "pourrais tu regarder", "tu peux aider stp", "merci de review",
            "approve this", "validation requise", "decision needed", "blocker",
            "incident security customer",
        ]
        for phrase in authored:
            with self.subTest(phrase=phrase):
                self.assertTrue(slack.candidate_hints(phrase, authored=True))
        for phrase in inbound:
            with self.subTest(phrase=phrase):
                self.assertTrue(slack.candidate_hints(phrase, authored=False))
        self.assertFalse(slack.candidate_hints("ordinary status update", authored=True))
        self.assertEqual(slack.candidate_hints("ordinary inbound mention", False, True), ["direct_mention"])

    def test_real_search_schema_variants(self):
        nested_user = match("1786185000.000", "I will own this")
        nested_user.pop("user_id")
        nested_user["user"] = {"id": slack.DEFAULT_USER, "name": "must-not-leak"}
        string_channel = match("1786185100.000", "resolved")
        string_channel["channel"] = "C12345678"
        opener = QueueOpener([search_payload([nested_user, string_channel]), search_payload([])])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["complete"])
        self.assertEqual(result["coverage"]["returned"], 2)
        self.assertEqual(result["events"][0]["author_id"], slack.DEFAULT_USER)
        self.assertNotIn("must-not-leak", json.dumps(result))

    def test_thread_context_is_paginated_bounded_around_target_and_sanitized(self):
        messages = []
        for i in range(20):
            ts = f"1786185{i:03d}.000"
            text = (" target " if i == 10 else " context ") + ("é" * 300)
            if i == 9:
                text += " " + XOXC + " password=hunter2"
            messages.append({"ts": ts, "text": text, "user": "U11111111"})
        opener = QueueOpener([
            {"ok": True, "messages": messages[:11], "has_more": True,
             "response_metadata": {"next_cursor": "NEXT"}},
            {"ok": True, "messages": messages[11:], "has_more": False,
             "response_metadata": {"next_cursor": ""}},
        ])
        result = slack.collect_thread_context(
            slack.SlackClient(CREDS, opener), CREDS, "C12345678", "1786185000.000", "1786185010.000"
        )
        coverage = result["coverage"]
        self.assertTrue(coverage["complete"])
        self.assertTrue(coverage["pagination_complete"])
        self.assertTrue(coverage["target_found"])
        self.assertEqual(coverage["pages"], 2)
        self.assertLessEqual(len(result["context"]["messages"]), 12)
        self.assertTrue(any(item["target"] for item in result["context"]["messages"]))
        self.assertTrue(all(len(item["snippet"]) <= 240 for item in result["context"]["messages"]))
        serialized = json.dumps(result)
        self.assertNotIn(XOXC, serialized)
        self.assertNotIn("hunter2", serialized)
        self.assertIn("cursor=NEXT", opener.requests[1].full_url)

    def test_thread_context_incomplete_pagination_fails_closed(self):
        opener = QueueOpener([{
            "ok": True, "messages": [{"ts": "1786185000.000", "text": "target"}],
            "has_more": True, "response_metadata": {"next_cursor": ""},
        }])
        result = slack.collect_thread_context(
            slack.SlackClient(CREDS, opener), CREDS, "C12345678", "1786185000.000", "1786185000.000"
        )
        self.assertTrue(result["coverage"]["fail_closed"])
        self.assertFalse(result["coverage"]["pagination_complete"])
        self.assertEqual(result["coverage"]["errors"], ["context_pagination_incomplete"])
        self.assertEqual(result["context"]["messages"], [])

    def test_delta_does_not_eagerly_fetch_thread_context(self):
        opener = QueueOpener([
            search_payload([match("1786185000.000", "I will handle this", thread_ts="1786184000.000")]),
            search_payload([]),
        ])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["complete"])
        self.assertEqual(len(opener.requests), 2)
        self.assertTrue(all("search.messages" in request.full_url for request in opener.requests))

    def test_api_error_is_fail_closed_and_does_not_leak_body_or_secrets(self):
        error = urllib.error.HTTPError(
            "https://slack.com/api/search.messages", 429, "rate limited " + XOXC, {}, io.BytesIO(XOXD.encode())
        )
        opener = QueueOpener([error])
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertFalse(result["coverage"]["complete"])
        self.assertTrue(result["coverage"]["incomplete"])
        self.assertTrue(result["coverage"]["fail_closed"])
        self.assertEqual(result["events"], [])
        serialized = json.dumps(result)
        self.assertNotIn(XOXC, serialized)
        self.assertNotIn(XOXD, serialized)
        self.assertEqual(result["coverage"]["errors"], ["http_error"])

    def test_page_cap_is_truncated_incomplete_and_fail_closed(self):
        payloads = [search_payload([], cursor=f"cursor-{i}") for i in range(slack.MAX_PAGES_PER_VIEW)]
        opener = QueueOpener(payloads)
        result = slack.collect(slack.SlackClient(CREDS, opener), CREDS, START, END, slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["truncated"])
        self.assertTrue(result["coverage"]["incomplete"])
        self.assertTrue(result["coverage"]["fail_closed"])
        self.assertEqual(result["coverage"]["errors"], ["page_cap_reached"])
        self.assertEqual(result["events"], [])

    def test_probe_returns_metadata_only(self):
        private = {"ok": True, "user": "private user", "team": "private workspace", "url": "private URL"}
        result = slack.probe(slack.SlackClient(CREDS, QueueOpener([private])), slack.DEFAULT_USER)
        self.assertTrue(result["coverage"]["authenticated"])
        self.assertTrue(result["coverage"]["complete"])
        self.assertEqual(result["events"], [])
        serialized = json.dumps(result)
        self.assertNotIn("private user", serialized)
        self.assertNotIn("private workspace", serialized)
        self.assertNotIn("private URL", serialized)

    def test_credentials_loaded_from_protected_file(self):
        import tempfile
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / ".env"
            path.write_text(f"SLACK_MCP_XOXC_TOKEN={XOXC}\nSLACK_MCP_XOXD_TOKEN='{XOXD}'\n")
            path.chmod(0o600)
            creds = slack.load_credentials(path)
            self.assertEqual(creds, CREDS)
            path.chmod(0o644)
            with self.assertRaises(slack.SafeError) as caught:
                slack.load_credentials(path)
            self.assertEqual(caught.exception.code, "credential_file_permissions")

    def test_main_emits_one_compact_json_document_on_failure(self):
        output = io.StringIO()
        with mock.patch.object(sys, "stdout", output):
            code = slack.main(["--start", "not-an-iso", "--end", "2026-08-08T11:00:00Z"])
        self.assertEqual(code, 1)
        lines = output.getvalue().splitlines()
        self.assertEqual(len(lines), 1)
        parsed = json.loads(lines[0])
        self.assertTrue(parsed["coverage"]["fail_closed"])
        self.assertEqual(parsed["events"], [])
        self.assertNotIn(XOXC, output.getvalue())
        self.assertNotIn(XOXD, output.getvalue())

    def test_unexpected_error_is_json_fail_closed_without_exception_text(self):
        leaked = "unexpected " + XOXC + " " + XOXD
        output = io.StringIO()
        with mock.patch.object(slack, "load_credentials", return_value=CREDS), mock.patch.object(
            slack, "probe", side_effect=RuntimeError(leaked)
        ), mock.patch.object(sys, "stdout", output):
            code = slack.main(["--probe"])
        emitted = output.getvalue()
        self.assertEqual(code, 1)
        parsed = json.loads(emitted)
        self.assertEqual(parsed["coverage"]["errors"], ["internal_error"])
        self.assertTrue(parsed["coverage"]["fail_closed"])
        self.assertNotIn(XOXC, emitted)
        self.assertNotIn(XOXD, emitted)


if __name__ == "__main__":
    unittest.main(verbosity=2)
