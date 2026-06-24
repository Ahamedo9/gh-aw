## 2026-06-22 - Argument Injection Defense-in-Depth
**Vulnerability:** Potential argument injection in `docker`, `pip`, and `uv` validation commands.
**Learning:** While `exec.Command` prevents shell injection, CLI tools can still misinterpret arguments starting with `-` as flags. The codebase had inconsistent use of the `--` delimiter to separate flags from positional arguments.
**Prevention:** Always use the `--` delimiter before user-controlled positional arguments in `exec.Command` calls, even if prefix validation (rejecting `-`) is already in place.
