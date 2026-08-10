"""End-to-end tests for scripts/tars-report-thread (stdlib only).

Run from the repo root:  python3 -m unittest discover -s tests -v

Each test drives the real script as a subprocess against a temp ~/.hermes tree,
a local fake Slack HTTP server and a stub `hermes` binary that records its argv.
`run_script` asserts the invariants that hold for every case: stdout stays empty
(the scheduler's silence marker), the bot token never leaks into stderr, and the
only Slack call ever made is a conversations.history GET (the reconciler posts
nothing).
"""

import fcntl
import json
import os
import subprocess
import sys
import threading
import unittest
import urllib.parse
from datetime import datetime, timedelta
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from tempfile import TemporaryDirectory
from zoneinfo import ZoneInfo

REPO = Path(__file__).resolve().parent.parent
SCRIPT = REPO / "scripts" / "tars-report-thread"
CHANNEL = "C0TESTCHAN0"
BOT = "U0TESTBOT00"
HUMAN = "U08BDJAMSRZ"
JOB_IDS = ["e231e5faf180", "62e8cd9db637", "759e08c598e3"]
KB_JOB = "3eb53a322510"
PARIS = ZoneInfo("Europe/Paris")
TOKEN = "test-token"


def paris_now():
    return datetime.now(PARIS)


def today():
    return paris_now().date().isoformat()


class FakeSlack:
    """Records every request; serves one canned conversations.history payload."""

    def __init__(self, messages=(), error=None):
        self.requests = []  # (method, path)
        self.auth_headers = []
        self.payload = ({"ok": False, "error": error} if error
                        else {"ok": True, "messages": list(messages), "has_more": False})
        recorder = self

        class Handler(BaseHTTPRequestHandler):
            def _serve(self):
                recorder.requests.append((self.command, self.path))
                recorder.auth_headers.append(self.headers.get("Authorization"))
                body = json.dumps(recorder.payload).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)

            do_GET = _serve
            do_POST = _serve

            def log_message(self, *args):
                pass

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()

    @property
    def base(self):
        return "http://127.0.0.1:%d" % self.server.server_address[1]

    def stop(self):
        self.server.shutdown()
        self.server.server_close()

    def query(self, index=0):
        path = self.requests[index][1]
        return urllib.parse.parse_qs(urllib.parse.urlparse(path).query)


class ReconcilerTest(unittest.TestCase):
    def setUp(self):
        tmp = TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.tmp = Path(tmp.name)
        self.home = self.tmp / "hermes"
        (self.home / "state").mkdir(parents=True)
        (self.home / "cron").mkdir()
        self.calls_path = self.tmp / "hermes-calls.log"
        self.state_path = self.home / "state" / "daily-report-thread.json"
        self.hermes_bin = self.make_hermes_stub()
        self.slack = None

    def tearDown(self):
        if self.slack:
            self.slack.stop()

    # --- fixtures -----------------------------------------------------------

    def make_hermes_stub(self, rc=0):
        path = self.tmp / "hermes-stub"
        path.write_text(
            "#!/bin/sh\n"
            'printf "%%s\\n" "$*" >> %s\n'
            '[ %d -eq 0 ] || echo "hermes stub: simulated failure" >&2\n'
            "exit %d\n" % (self.calls_path, rc, rc)
        )
        path.chmod(0o755)
        return path

    def write_jobs(self, deliver):
        """deliver: one string for all three report jobs, or {job_id: value}."""
        if isinstance(deliver, str):
            deliver = {jid: deliver for jid in JOB_IDS}
        jobs = [
            {"id": JOB_IDS[0], "name": "Gaetan daily work brief",
             "prompt": "Prepare Gaetan's morning daily ...",
             "skills": ["daily-work-brief"], "skill": "daily-work-brief",
             "schedule": {"kind": "cron", "expr": "30 8 * * 1-5", "display": "30 8 * * 1-5"},
             "repeat": {"times": None, "completed": 1},
             "enabled": True, "state": "scheduled", "last_status": "ok",
             "deliver": deliver[JOB_IDS[0]], "no_agent": False,
             "script": None, "workdir": None},
            {"id": JOB_IDS[1], "name": "Gaetan engagement checker",
             "prompt": "Run `engagement-checker` end to end ...",
             "skills": ["engagement-checker"], "skill": "engagement-checker",
             "schedule": {"kind": "cron", "expr": "*/30 10-16 * * 1-5"},
             "repeat": {"times": None, "completed": 8},
             "enabled": True, "state": "scheduled", "last_status": "ok",
             "deliver": deliver[JOB_IDS[1]], "no_agent": False,
             "script": None, "workdir": None},
            {"id": JOB_IDS[2], "name": "Gaetan engagement checker final pass",
             "prompt": "Run `engagement-checker` end to end ...",
             "skills": ["engagement-checker"], "skill": "engagement-checker",
             "schedule": {"kind": "cron", "expr": "0 17 * * 1-5"},
             "repeat": {"times": None, "completed": 0},
             "enabled": True, "state": "scheduled", "last_status": None,
             "deliver": deliver[JOB_IDS[2]], "no_agent": False,
             "script": None, "workdir": None},
            # unrelated job sharing the channel — must never be edited
            {"id": KB_JOB, "name": "mc-metarepo-refresh", "script": "kb-refresh.sh",
             "no_agent": True,
             "schedule": {"kind": "interval", "minutes": 60, "display": "every 60m"},
             "repeat": {"times": None, "completed": 1},
             "enabled": True, "state": "scheduled", "last_status": "ok",
             "deliver": "slack:%s:1786359613.979759" % CHANNEL, "workdir": None},
        ]
        (self.home / "cron" / "jobs.json").write_text(json.dumps(jobs, indent=2))

    def write_state(self, date, parent_ts):
        self.state_path.write_text(json.dumps({
            "channel": CHANNEL, "date": date, "parent_ts": parent_ts,
            "updated_at": "2026-08-09T07:25:00+02:00",
        }))

    # --- helpers ------------------------------------------------------------

    def run_script(self, *extra, expect_rc=0, env_token=True):
        argv = [sys.executable, str(SCRIPT),
                "--channel", CHANNEL,
                "--bot-user", BOT,
                "--jobs", ",".join(JOB_IDS),
                "--hermes-home", str(self.home),
                "--hermes-bin", str(self.hermes_bin),
                "--tz", "Europe/Paris"]
        if self.slack:
            argv += ["--slack-base", self.slack.base]
        argv += list(extra)
        env = dict(os.environ, SLACK_BOT_TOKEN=TOKEN)
        if not env_token:
            env.pop("SLACK_BOT_TOKEN", None)
        done = subprocess.run(argv, capture_output=True, text=True, env=env)
        self.assertEqual(done.stdout, "", "stdout must stay empty (silence marker)")
        self.assertNotIn(TOKEN, done.stderr)
        self.assertEqual(done.returncode, expect_rc, done.stderr)
        if self.slack:
            for method, path in self.slack.requests:
                self.assertEqual(method, "GET", "the reconciler must never post")
                self.assertTrue(path.startswith("/conversations.history?"), path)
        return done

    def hermes_calls(self):
        if not self.calls_path.exists():
            return []
        return self.calls_path.read_text().splitlines()

    def expected_edits(self, deliver, job_ids=JOB_IDS):
        return sorted("cron edit %s --deliver %s" % (jid, deliver) for jid in job_ids)

    def read_state(self):
        return json.loads(self.state_path.read_text())

    # --- cases --------------------------------------------------------------

    def test_1_no_parent_today_flips_jobs_top_level(self):
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s:1786359613.979759" % CHANNEL)
        self.run_script()
        self.assertEqual(sorted(self.hermes_calls()),
                         self.expected_edits("slack:%s" % CHANNEL))
        state = self.read_state()
        self.assertEqual(state["channel"], CHANNEL)
        self.assertEqual(state["date"], today())
        self.assertIsNone(state["parent_ts"])
        self.assertIn("+", state["updated_at"], "updated_at must carry a tz offset")

    def test_2_earliest_bot_top_level_message_wins(self):
        parent = "1786370000.000100"
        self.slack = FakeSlack(messages=[  # Slack returns newest first
            {"ts": "1786400000.000100", "thread_ts": "1786400000.000100",
             "user": HUMAN, "text": "human top-level decoy"},
            {"ts": "1786399000.000100", "thread_ts": "1786360000.000100",
             "user": BOT, "text": "bot threaded reply decoy"},
            {"ts": "1786398000.000100", "user": BOT, "text": "later bot top-level"},
            {"ts": parent, "thread_ts": parent, "user": BOT, "text": "the day's brief"},
            {"ts": "1786369000.000100", "user": BOT, "subtype": "channel_join"},
            {"ts": "1786368000.000100", "user": HUMAN, "subtype": "channel_join"},
        ])
        self.write_jobs("slack:%s" % CHANNEL)
        self.run_script()
        self.assertEqual(sorted(self.hermes_calls()),
                         self.expected_edits("slack:%s:%s" % (CHANNEL, parent)))
        self.assertEqual(self.read_state()["parent_ts"], parent)

    def test_3_fast_path_makes_no_http_and_no_edits(self):
        parent = "1786370000.000100"
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s:%s" % (CHANNEL, parent))
        self.write_state(today(), parent)
        self.run_script()
        self.assertEqual(self.slack.requests, [], "fast path must not call Slack")
        self.assertEqual(self.hermes_calls(), [])

    def test_4_reconciles_after_state_loss(self):
        parent = "1786370000.000100"
        self.slack = FakeSlack(messages=[
            {"ts": parent, "user": BOT, "text": "today's brief, already sent"},
        ])
        # state write failed after the send: no state file, jobs still on yesterday's
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        self.assertFalse(self.state_path.exists())
        self.run_script()
        self.assertEqual(sorted(self.hermes_calls()),
                         self.expected_edits("slack:%s:%s" % (CHANNEL, parent)))
        self.assertEqual(self.read_state()["parent_ts"], parent)

    def test_5_paris_rollover_queries_today_midnight(self):
        self.slack = FakeSlack(messages=[])
        yesterday = (paris_now().date() - timedelta(days=1)).isoformat()
        self.write_state(yesterday, "1786280000.000100")
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        self.run_script()
        expected = paris_now().replace(hour=0, minute=0, second=0,
                                       microsecond=0).timestamp()
        query = self.slack.query()
        self.assertAlmostEqual(float(query["oldest"][0]), expected, places=3)
        self.assertEqual(query["channel"], [CHANNEL])
        self.assertEqual(query["limit"], ["999"])
        self.assertEqual(query["inclusive"], ["true"])
        self.assertEqual(sorted(self.hermes_calls()),
                         self.expected_edits("slack:%s" % CHANNEL))
        self.assertIsNone(self.read_state()["parent_ts"])

    def test_6_idempotent_when_jobs_already_correct(self):
        parent = "1786370000.000100"
        self.slack = FakeSlack(messages=[{"ts": parent, "user": BOT, "text": "brief"}])
        self.write_jobs("slack:%s:%s" % (CHANNEL, parent))  # already right, no state
        self.run_script()
        self.assertEqual(self.hermes_calls(), [])
        self.assertEqual(self.read_state()["parent_ts"], parent)

    def test_7_cron_edit_failure_fails_and_leaves_state_untouched(self):
        self.hermes_bin = self.make_hermes_stub(rc=1)
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        done = self.run_script(expect_rc=1)
        self.assertIn("cron edit", done.stderr)
        self.assertEqual(len(self.hermes_calls()), 3, "every job is still attempted")
        self.assertFalse(self.state_path.exists(), "state must not advance on failure")

    def test_8_dry_run_changes_nothing(self):
        self.slack = FakeSlack(messages=[{"ts": "1786370000.000100", "user": BOT}])
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        done = self.run_script("--dry-run")
        self.assertIn("dry-run", done.stderr)
        self.assertEqual(self.hermes_calls(), [])
        self.assertFalse(self.state_path.exists())

    def test_9_missing_job_id_is_a_failure(self):
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s" % CHANNEL)
        jobs_path = self.home / "cron" / "jobs.json"
        jobs = [j for j in json.loads(jobs_path.read_text()) if j["id"] != JOB_IDS[1]]
        jobs_path.write_text(json.dumps(jobs))
        done = self.run_script(expect_rc=1)
        self.assertIn(JOB_IDS[1], done.stderr)
        self.assertEqual(self.hermes_calls(), [])
        self.assertFalse(self.state_path.exists())

    def test_10_lock_held_elsewhere_is_a_no_op(self):
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        lock = os.open(self.home / "state" / "daily-report-thread.lock",
                       os.O_CREAT | os.O_RDWR, 0o600)
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        try:
            done = self.run_script()
        finally:
            os.close(lock)
        self.assertIn("lock held", done.stderr)
        self.assertEqual(self.slack.requests, [])
        self.assertEqual(self.hermes_calls(), [])
        self.assertFalse(self.state_path.exists())

    def test_11_slack_error_fails_without_touching_jobs(self):
        self.slack = FakeSlack(error="channel_not_found")
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        done = self.run_script(expect_rc=1)
        self.assertIn("channel_not_found", done.stderr)
        self.assertEqual(self.hermes_calls(), [])
        self.assertFalse(self.state_path.exists())

    def test_12_env_file_fallback_supplies_token(self):
        # production path: the scheduler's sanitized env strips SLACK_BOT_TOKEN,
        # so the script must find it in <hermes-home>/.env
        self.slack = FakeSlack(messages=[])
        self.write_jobs("slack:%s:1786280000.000100" % CHANNEL)
        (self.home / ".env").write_text(
            "OTHER_VAR=whatever\nexport SLACK_BOT_TOKEN='%s'\n" % TOKEN)
        self.run_script(env_token=False)
        self.assertEqual(self.slack.auth_headers, ["Bearer %s" % TOKEN])
        self.assertEqual(sorted(self.hermes_calls()),
                         self.expected_edits("slack:%s" % CHANNEL))

    def test_13_corrupt_state_self_heals_via_discovery(self):
        parent = "1786370000.000100"
        self.slack = FakeSlack(messages=[{"ts": parent, "user": BOT, "text": "brief"}])
        self.write_jobs("slack:%s" % CHANNEL)
        self.state_path.write_text("{not json")
        self.run_script()
        self.assertEqual(self.read_state()["parent_ts"], parent)


if __name__ == "__main__":
    unittest.main()
