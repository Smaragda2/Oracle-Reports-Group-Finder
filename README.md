# Oracle Report Group Finder

A small, portable Windows utility for finding every **Frame** and **Repeating Frame** in an Oracle Reports layout whose **Source** points to a selected Data Model **Group**.

It is especially useful with older or large Oracle Reports where object names such as `F_154`, `F_23`, `R_17`, etc. do not make the relationship between the layout and a group obvious.

## What it does

Given a group such as:

```text
G_PERCENTAGE1
```

the tool searches the report definition and shows all layout objects whose `Source` uses that group, for example:

```text
GROUP: G_PERCENTAGE1

Frame: F_154
  Parent Frames:
    - F_20

Frame: F_23

Repeating Frame: R_10
  Child Frames:
    - F_88
```

The results can include:

- Frame name
- Object type: `Frame` or `Repeating Frame`
- Matching Source Group
- Parent Frames
- Child Frames
- Layout Section, when available

This gives Oracle Reports developers a practical **Find Usages for Group/Source** workflow that is missing from the classic Reports Builder UI.

## Typical use cases

- Find where a Data Model Group is used in the report layout.
- Locate Frames whose names do not match the Group name.
- Understand old reports with generated names such as `F_154`, `F_23`, etc.
- Investigate layout ownership before changing or migrating an Oracle Report.
- Assist Oracle Reports / Oracle Forms modernization and migration work.
- Quickly inspect report structure without checking each Property Palette entry manually.

## Supported input files

| File | Supported | Converter required? | Notes |
|---|---:|---:|---|
| `.rdf` | Yes | Yes | Converted internally to REX before parsing. |
| `.rex` | Yes | No | Parsed directly. |
| `.xml` | Yes | No | Parsed directly. |

### RDF conversion

For `.rdf` files, Oracle's Reports Converter is still required because RDF is Oracle's binary report-definition format.

The tool accepts converter executables/scripts such as:

```text
RWCON60.EXE
rwconverter.exe
rwconverter.bat
rwconverter.cmd
```

For Oracle Reports 6i, `RWCON60.EXE` is the normal choice. The tool converts:

```text
RDF -> REX
```

using parameters equivalent to:

```text
stype=rdffile
source=<temporary RDF>
dtype=rexfile
dest=<temporary REX>
overwrite=yes
batch=yes
logfile=<log file>
```

If `USERID` is entered in the UI, it is also passed to the Oracle converter.

The original RDF is not modified.

## Requirements

### When opening an RDF

You need:

1. `OracleReportGroupFinder.exe`
2. The `.rdf` file to inspect
3. A compatible Oracle Reports converter, for example `RWCON60.EXE` for Oracle Reports 6i

No Java/JRE installation is required.
No .NET installation is required.
No installer is required.

> The Oracle converter is **not distributed with this repository**. It is an Oracle component and must come from an Oracle Reports installation/environment that you are licensed to use.

### When opening REX or XML

You only need:

1. `OracleReportGroupFinder.exe`
2. The `.rex` or `.xml` file

No Oracle converter is required for direct REX/XML loading.

## Windows compatibility

The published executable is **64-bit Windows**.

### Verified

- Windows Server 2008 R2 x64 — tested successfully with the tool.

### Compatibility target

The executable/source is designed for:

- Windows Server 2008 R2 x64 or later
- Windows Server 2012 / 2012 R2
- Windows Server 2016
- Windows Server 2019
- Windows Server 2022
- Windows Server 2025
- Windows 7 x64 or later
- Windows 8 / 8.1 x64
- Windows 10 x64
- Windows 11 x64

The compatibility build uses **Go 1.20.x**, because Go 1.20 is the final upstream Go release line that supports Windows 7 and Windows Server 2008 R2.

## Oracle Reports compatibility

The tool was built specifically around the report-definition formats used by classic Oracle Reports and has working support for the Oracle Reports 6i `RWCON60.EXE` workflow.

It can also accept later `rwconverter` executables/scripts as long as they support RDF-to-REX conversion and can read the source RDF.

Important: the Oracle converter itself must be compatible with the RDF you are trying to convert. This tool does not replace Oracle's RDF reader; it automates the conversion and then analyzes the resulting definition.

## How to use

1. Run `OracleReportGroupFinder.exe`.
2. If you are opening an `.rdf`, select the Oracle Reports converter, e.g. `RWCON60.EXE`.
3. Select the report file: `.rdf`, `.rex`, or `.xml`.
4. Optionally enter `USERID` if your Oracle conversion requires it.
5. Click **Convert & Load**.
6. Select or type the Group, e.g. `G_PERCENTAGE1`.
7. Click **Search Group**.
8. Review all Frames and Repeating Frames whose Source matches the selected Group.

For RDF input, the generated REX is normally saved next to the original report as:

```text
<report-name>_groupfinder.rex
```

If the original report directory is not writable, the tool can continue using its temporary converted file.

## Non-ASCII / Greek paths

Older Oracle Reports tools can have problems with Unicode or non-ASCII paths.

To reduce this problem, the application copies the RDF to a temporary **ASCII-only** working directory before calling the Oracle converter.

For example, an RDF stored under a path containing Greek characters can still be converted without passing that original path directly to `RWCON60.EXE`.

## How the parsing works

### XML

The XML parser reads the report hierarchy and records:

- Groups
- Frames
- Repeating Frames
- Source properties
- Parent/child relationships

### REX

The REX parser reads Oracle `DEFINE` blocks and known properties used for Groups and layout Frames.

When an explicit parent-frame identifier exists, it is used directly.

Older REX variants do not always expose a documented parent relationship. In that case, the tool performs a best-effort hierarchy reconstruction using the layout rectangles: the smallest enclosing Frame is treated as the parent.

The **Source match itself** is independent from this visual parent/child inference. The hierarchy is supplementary information.

## Generated files

Depending on usage, the tool may create:

```text
OracleReportGroupFinder.ini
<report-name>_groupfinder.rex
```

`OracleReportGroupFinder.ini` stores only the last selected converter path, when the executable directory is writable.

Temporary conversion files are created under the Windows temporary directory and normally removed automatically.

## Privacy / network behavior

The application works locally.

- It does not upload report files.
- It does not call an online API.
- It does not require an internet connection.
- Report conversion is performed locally by the selected Oracle converter.

## Limitations

- RDF cannot be parsed directly; an Oracle Reports converter is required first.
- The published executable is x64 only.
- REX layout hierarchy may require geometric inference when explicit ownership information is unavailable.
- Different Oracle Reports releases may produce slightly different REX structures. If a specific report format is not recognized, the REX parser may need another mapping for that release.
- `USERID` is passed to the Oracle converter when provided; it is not stored by this tool.

## Build from source

The application is written in Go and uses only the Go standard library plus native Win32 APIs. There are no third-party Go dependencies.

### Build the Windows Server 2008 R2-compatible executable

Install **Go 1.20.x** and run:

```bat
build.bat
```

The executable will be created at:

```text
dist\OracleReportGroupFinder.exe
```

`build.bat` intentionally rejects Go 1.21+ for the legacy-compatible build because upstream Go 1.21 dropped support for Windows 7 and Windows Server 2008 R2.

### GitHub Actions

The repository includes:

```text
.github/workflows/build-windows.yml
```

It builds the application with Go `1.20.14` and uploads the Windows x64 executable as a workflow artifact. This allows the project to be built directly in GitHub without an IDE.

## Repository structure

```text
.
├── .github/
│   └── workflows/
│       └── build-windows.yml
├── .gitignore
├── README.md
├── build.bat
├── go.mod
└── main.go
```

## GitHub Releases

A good repository setup is to keep source code in the repository and attach the compiled `.exe` to **GitHub Releases** instead of committing binaries to the source tree.

Example release asset:

```text
OracleReportGroupFinder.exe
```

## Why this tool exists

Oracle Reports Builder exposes the `Source` property in the Property Palette, but large legacy reports can contain many Frames with non-descriptive generated names. There is no convenient modern-style `Find Usages` workflow that starts from a Group and immediately lists every layout Frame using it.

Oracle Report Group Finder fills that gap: select a Group and get the relevant Frames and Repeating Frames in one place, making report maintenance, debugging, reverse engineering, and migration much faster.

## Project status / disclaimer

This is an independent utility and is not affiliated with or endorsed by Oracle. Oracle, Oracle Reports, and related product names are trademarks of Oracle and/or its affiliates.
