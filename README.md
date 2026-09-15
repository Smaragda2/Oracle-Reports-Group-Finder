# Oracle Report Group Finder

A portable Windows utility for finding every **Frame** and **Repeating Frame** in an Oracle Reports layout whose **Source** points to a selected Data Model **Group**.

It is especially useful with older or large Oracle Reports where layout objects have non-descriptive names such as `F_154`, `F_23`, `R_17`, etc., making it difficult to understand which objects belong to a specific group through Oracle Reports Builder alone.

## What it does

Given a Data Model group such as:

```text
G_PERCENTAGE1
```

the tool searches the report definition and shows all supported layout objects whose `Source` uses that group, for example:

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

This provides a practical **Find Usages for Group/Source** workflow that is missing from the classic Oracle Reports Builder UI.

## Typical use cases

- Find where a Data Model Group is used in the report layout.
- Locate Frames whose names do not match the Group name.
- Understand legacy reports with generated names such as `F_154`, `F_23`, etc.
- Investigate layout ownership before changing an Oracle Report.
- Assist Oracle Reports / Oracle Forms modernization and migration work.
- Quickly inspect report structure without checking each Property Palette entry manually.

## Download

The ready-to-run Windows executable is published in the repository's **Releases** section:

```text
OracleReportGroupFinder.exe
```

No installation is required. The application is portable and can be run directly.

## Supported input files

| File | Supported | Converter required? | Notes |
|---|---:|---:|---|
| `.rdf` | Yes | Yes | Converted internally to REX before parsing. |
| `.rex` | Yes | No | Parsed directly. |
| `.xml` | Yes | No | Parsed directly. |

## Requirements

### When opening an RDF

You need:

1. `OracleReportGroupFinder.exe`
2. The `.rdf` report file
3. A compatible Oracle Reports converter, for example `RWCON60.EXE` for Oracle Reports 6i

Supported converter names include:

```text
RWCON60.EXE
rwconverter.exe
rwconverter.bat
rwconverter.cmd
```

For Oracle Reports 6i, `RWCON60.EXE` is the normal choice.

The tool performs the conversion:

```text
RDF -> REX
```

and then analyzes the generated report definition.

If `USERID` is entered in the application, it is passed to the Oracle converter when required by the report/environment.

The original RDF file is not modified.

> The Oracle converter is **not distributed with this repository**. It is an Oracle component and must come from an Oracle Reports installation/environment that you are licensed to use.

### When opening REX or XML

You only need:

1. `OracleReportGroupFinder.exe`
2. The `.rex` or `.xml` report file

No Oracle Reports converter is required when loading REX or XML directly.

## No additional runtime installation required

The published executable does **not** require:

- Java / JRE
- .NET installation
- An installer
- An internet connection

For `.rdf` input, only the Oracle Reports converter described above is additionally required.

## Windows compatibility

The published executable is **64-bit Windows**.

### Verified

- **Windows Server 2008 R2 x64** — tested successfully.

### Compatible target environments

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

## Oracle Reports compatibility

The tool was created primarily for classic Oracle Reports and has been tested with the **Oracle Reports 6i `RWCON60.EXE`** workflow.

It can also use later `rwconverter` executables/scripts when they support RDF-to-REX conversion and can read the source RDF.

The Oracle converter itself must be compatible with the RDF being converted. Oracle Report Group Finder does not replace Oracle's RDF reader; it automates the conversion and analyzes the resulting report definition.

## How to use

1. Run `OracleReportGroupFinder.exe`.
2. If opening an `.rdf`, select the Oracle Reports converter, for example `RWCON60.EXE`.
3. Select the report file: `.rdf`, `.rex`, or `.xml`.
4. Optionally enter `USERID` if the Oracle conversion requires it.
5. Click **Convert & Load**.
6. Select or type the Group, for example `G_PERCENTAGE1`.
7. Click **Search Group**.
8. Review all Frames and Repeating Frames whose Source matches the selected Group.

For RDF input, the generated REX is normally saved next to the original report as:

```text
<report-name>_groupfinder.rex
```

If the original report directory is not writable, the application can continue using the temporary converted file.

## Non-ASCII / Greek paths

Older Oracle Reports utilities may have problems with Unicode or non-ASCII paths.

To reduce this problem, Oracle Report Group Finder copies the RDF to a temporary **ASCII-only** working directory before calling the Oracle converter.

This allows reports located in folders containing Greek or other non-ASCII characters to be processed without passing the original path directly to older Oracle utilities such as `RWCON60.EXE`.

## How the analysis works

### XML

The XML parser reads the report hierarchy and records relevant information including:

- Groups
- Frames
- Repeating Frames
- Source properties
- Parent/child relationships

### REX

The REX parser reads Oracle report-definition structures and properties used for Groups and layout Frames.

When an explicit parent-frame relationship is available, it is used directly.

Some older REX variants do not expose a clear documented parent relationship. In those cases, the tool can perform a best-effort hierarchy reconstruction using layout rectangles, where the smallest enclosing Frame is treated as the parent.

The **Source match itself does not depend on this hierarchy inference**. Parent/child information is supplementary to the Group-to-Source match.

## Generated files

Depending on usage, the tool may create:

```text
OracleReportGroupFinder.ini
<report-name>_groupfinder.rex
```

`OracleReportGroupFinder.ini` stores the last selected converter path when the executable directory is writable.

Temporary conversion files are created under the Windows temporary directory and are normally removed automatically.

## Privacy and network behavior

The application works locally.

- Report files are not uploaded anywhere.
- No online API is called.
- No internet connection is required.
- RDF conversion is performed locally by the selected Oracle Reports converter.

## Limitations

- Binary RDF files cannot be analyzed directly by this tool; they must first be converted by a compatible Oracle Reports converter.
- The published executable is x64 only.
- REX layout hierarchy may require geometric inference when explicit ownership information is unavailable.
- Different Oracle Reports releases may produce slightly different REX structures. A specific format variant may require additional parser mapping.
- `USERID`, when provided, is passed to the Oracle converter and is not stored by Oracle Report Group Finder.

## Source code

The repository includes the application source code for reference, maintenance and future improvements.

The supported executable for normal use is the version published under **GitHub Releases**.

## Disclaimer

Oracle and Oracle Reports are trademarks of Oracle Corporation and/or its affiliates.

Oracle Report Group Finder is an independent utility and is not affiliated with or endorsed by Oracle Corporation. Oracle binaries such as `RWCON60.EXE` or `rwconverter` are not included in this repository.
