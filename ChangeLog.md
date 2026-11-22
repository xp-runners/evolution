XP Runners change log
=====================

## ?.?.? / ????-??-??

* Ensured the list of PHP extensions is unique, loading them multiple times
  will raise warnings
  (@thekid)

## 10.0.0 / 2026-11-22

* Fixed issue #2: Startup warnings (caused by quoting command line args)
  (@thekid)
* Ported XP runners to GoLang, creating small self-contained binaries
  without dependencies on .NET or Mono runtimes. For the C# version,
  see https://github.com/xp-runners/reference
  (@thekid, GPT-6.2-Sol)
