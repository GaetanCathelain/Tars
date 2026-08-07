# Project pitch — Tars (verbatim, 2026-08-07)

Okay, big project. I want to create my personnal pro dedicated Hermes agent assistant.

Get to know what Hermes agent is (<https://hermes-agent.nousresearch.com/>).

The interaction plane I want: Slack. Tars directely integrated in Slack, only responds to me, has
access to all the channels I have access to.

Which Data/mcp it should have access to:

- Slack: Tars interaction
- Slack: My personal account (all DMs, ...) -> This might be a tricky one, you can do research on that
- Gmail (via Himalaya)
- Linear (personnal API token)
- GitHub (personnal Oauth)
- GitHub: the mc-metarepo. The mc-metarepo is our huge knowledge base
- Notion
- Google Calendar
- Other data that won't be "live", either via GBrain or pure ../mc-kestra dumps (this one is postponed
  until I finish the work on mc-kestra)

For all those services, minimal friction for authentication is a must. I can setup a browser logged in
with my gaetan.cathelain@mobile.club account as needed.

Which machines it should have access to (via SSH):

- This VM (cooper VM)
- My MacOS
- My personnal Hermes (to dump the meme creator skill for example)
- All machines on Tailscale pro account

Which apps it should have access to:

- Orca on Cooper
- Cloakbrowser for himself

Which Hermes plugins/skills it should have installed:

- RTK for lower token usage
- hindsight for local memory (local, basic, use hermes memory config to setup)
- hermes-lcm
- i-have-adhd

Other requirements:

- /sethome should be a DM conversation between me and him
- You can use "hermes chat" CLI command to talk to it directly, simulating commands with him, or using
  A2A protocol that was newly released in Hermes latest version.
- Tars will be the default profile of the Hermes instance

Slack setup:

- I previously created a Tars Slack app, it's used by another profile on my personal Hermes, you should
  delete that profile entirely on my personal Hermes VM and re-setup the same Slack app on this
  professional Hermes VM we're building.

Where it should run:

- A new VM on Proxmox called "Tars" with 8 CPU, 8GB of RAM, ubuntu-latest with GUI setup. 50 GB of
  disk.

The goal of Tars:

- Be my personnal assistant. I've been having issues with reporting and overall structuring and
  orchestrating my work, especially since I'm going so fast with AI now. I want to have a personal Kanban
  board somewhere I can follow, managed mostly by Tars and Tars should remind me of what I should do,
  prepare my dailys, create reports, orchestrate Orca on my Cooper VM to start working on tasks, ...
- Be a techical assistant, with access to almost everything I have access to profesionally and automate
  as much tasks as possible.
- Be able to orchestrate work on my Cooper VM that is running Orca, directly integrate with Orca to
  start/stop/orchestrate tasks idealy

What Tars should NOT BE:

- A coding god. It should have technical knowledge but never aim to implement anything itself. Any
  actual coding/fixing work should be done by the Cooper VM. No PR should ever be created by Tars

---

You can structure your work by using (and completely revamping)
<https://github.com/GaetanCathelain/Tars>. It's an old project with almost nothing on it, you can delete
every branch etc.

We should first start our plan by trying to fetch all required credentials, and test each, one by one,
so authentification doesn't become an issue later on.
