# Diagnosing anomalous Slack busy acknowledgments

Use this note when a new Slack top-level message receives a busy acknowledgment such as `Interrupting current task`, even though Slack threads are expected to have separate Hermes sessions.

## Evidence hierarchy

Do not infer session binding from the acknowledgment text alone. Establish these independently:

1. **Slack topology:** recover the message `ts`, `thread_ts`, channel, workspace, and event time from Slack.
2. **Computed routing:** inspect the actual session key. For threaded channel traffic it should include workspace, channel, and thread root; compare it with the allegedly interrupted run's key.
3. **Session state:** verify whether the routed session was fresh or had prior history.
4. **Busy path:** locate the exact code path that emits the acknowledgment and its precondition (normally, the computed key is already active).
5. **Event multiplicity:** inspect adapter registration and deduplication for overlapping Slack event paths such as `message` and `app_mention`, reconnect redelivery, or message edits.
6. **Runtime chronology:** correlate inbound logs, session creation, acknowledgment timestamp, and agent start. A duplicate intercepted by the adapter's busy guard may never reach the runner's normal inbound log, so one logged inbound event does not exclude a second adapter-level delivery.

## Interpretation

- If the new and old tasks have different verified session keys, the busy acknowledgment does **not** prove that separate threads share context.
- If the new session is fresh but receives a busy acknowledgment immediately after creation, duplicate delivery or a deduplication race is a plausible mechanism: the first event starts the session and the second sees that newly active key.
- Treat that mechanism as a hypothesis until raw event or debug-level evidence confirms both deliveries. Deduplication intended to suppress `message`/`app_mention` overlap makes an observed anomaly a defect candidate, not proof of the exact root cause.
- Report three separate verdicts: **binding behavior**, **observed anomaly**, and **causal confidence**. Say `unintended acknowledgment; precise cause unproven` when evidence establishes the first two but not the third.

## Common mistake

Do not defend the documented thread model by explaining `busy_input_mode` alone. First reconcile the contradictory concrete example. The correct answer may be that the binding model is intact while the acknowledgment is erroneous.
