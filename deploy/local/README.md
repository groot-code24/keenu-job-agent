# Local edition: starter job boards

`starter-boards/<provider>.json` is the job-board catalogue a fresh local install starts
from: a short list of tech companies on Greenhouse, Lever and Ashby, in the `{board,
company}` seed format `cmd/harvest-boards` reads.

It is a candidate list, not trusted data. On every `docker compose up` the `seed-boards`
service runs `harvest-boards --apply` over each file, which probes every slug against the
platform's own public API and adds only boards that answer with jobs. A wrong or retired
slug is dropped, and a board already in the catalogue is skipped, so later starts cost little.

The `crawler` service then crawls the catalogue on the scheduler's cadence
(`cmd/ingest-scheduler` with `INGEST_SCHEDULER_LAUNCHER=process`), and `search-drain`
pushes what it found into search.

To add a company: append it to the provider's file (the slug is the part of the board URL
after the host, e.g. `boards.greenhouse.io/<slug>`), or add it directly:

```sh
docker compose run --rm --entrypoint /app/add-board app --provider=greenhouse --board=<slug> --company="<Name>" --apply
```
