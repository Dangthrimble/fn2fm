# fn2fm - Filename to forScore Metadata

## forScore Metadata

By default, scores and bookmarks in forScore 14.0 can be tagged with the following [forScore metadata](https://forscore.co/documentation/metadata/):
- Title
- Rating
- Difficulty
- Reference
- Time (i.e. duration)
- Key
- Composer(s)
- Genre(s)
- Tag(s)
- Label(s)

 Some of these types might not fit individual needs and all bar "Title" can be renamed to better suit the needs of the individual.

 ## PDF Metadata

 PDF files can also be tagged with [forScore-specific PDF metadata](https://forscore.co/developers-pdf-metadata/), and these can be fetched by forScore and used to populate forScore metadata with the exception of "Reference" and "Label(s)", as follows:

### PDF Specification

| PDF Value | forScore Value | Format                                                                            |
| --------- | -------------- | --------------------------------------------------------------------------------- |
| Title     | Title          | Single value                                                                      |
| Author    | Composers      | One or more comma-separated values                                                |
| Subject   | Genres         | One or more comma-separated values                                                |
| Keywords  | Tags           | One or more comma-separated values (excluding keywords listed in the chart below) |

### Specialty Keywords

| Keyword      | forScore Value | Format                                 |
| ------------ | -------------- | -------------------------------------- |
| rating:N     | Rating         | Whole number between 0 and 5           |
| difficulty:N | Difficulty     | Whole number between 0 and 3           |
| duration:N   | Duration       | Non-negative whole number (in seconds) |
| keysf:N      | Key            | A whole number between -7 and 7        |
| keymi:N      |                | 0 (major) or 1 (minor)                 |

 The purpose of fn2fm is to allow a PDF filename to be constructed and tagged in such a way that, when parsed, PDF metadata can be derived and added to the file prior to it being imported into forScore. The additional benefit is that, if the file's naming convention is chosen well, it will reduce the likelihood of filename clashes (e.g. scores with the same name by different composers or arrangers; the same score in different keys) while keeping the filenames relatively short.

## Supported Character Set

An additional possible constraint on PDF filenames (e.g. for choirs) is that, where forScore files are shared via cloud storage, the supported character set has to comply with that of the cloud storage used, as well as that of the device(s) on which the PDF files are prepared for sharing (and on which fn2fm is run). Two popular cloud storage solutions are [Dropbox](https://www.dropbox.com) and [Box](https://www.box.com/) which both provide guidance on filenaming:
- [Naming Dropbox files and folders](https://help.dropbox.com/organize/file-names)
- [Troubleshooting Uploads to Box](https://support.box.com/hc/en-us/articles/360044196773-Troubleshooting-Uploads-to-Box)

Based on those two cloud storage solutions, the characters to avoid are:
- non-printable ASCII characters, which comprise the [ASCII Control Characters](https://www.ascii-code.com/characters/control-characters)
- " (double quote, quotation mark)
- \* (asterisk)
- . (period, full stop)
- / (forward slash, solidus)
- : (colon)
- < (less than)
- \> (greater than)
- ? (question mark)
- \ (backslash, reverse solidus)
- | (vertical bar or pipe)
- emojis or emoticons

In addition:
- Dropbox says to avoid special characters (without specifying what constitutes a special character) and doesn't support names with trailing spaces
- Box only supports the [Unicode Basic Multilingual Plane (BMP)](https://codepoints.net/basic_multilingual_plane) characters and doesn't support names with leading or trailing spaces

Semicolons (`;`) are also forbidden throughout the filename. Only `~` is supported as the metadata separator; semicolon-based filenames are rejected.

## fn2fm Score Naming Convention

The naming convention used by fn2fm has been developed to support the needs of our choir [Cambrensis](https://www.cambrensis.org.uk/), uses only [7-bit ASCII printable characters](https://www.ascii-code.com/characters/printable-characters) constrained by the characters to be avoided listed above, and supports the following metadata:
- Score Title (the same as the PDF filename, but without the ".pdf" extension)
- Composer(s)
- Arranger(s) (used by repurposing "Genre(s)")
- Initial key signature
- Whether or not the score includes an accompaniment (this is visible in the forScore "Tag(s)")

The filename is constructed from two parts separated by a tilde ("~"):
- The score title
- The score metadata

To minimise uncertainty, the filename must contain exactly one tilde. Filenames without a tilde are rejected. If there are two or more, fn2fm will not know which separates the score title from the metadata and the filename will be rejected.

The score metadata is formatted as follows and, apart from the limitation of no trailing spaces in the filename, each can have leading and trailing spaces if they aid readability (we just use a single space before and after the tilde):
- <composer(s)\>_<arranger(s)\>\[<initialKeySignature\>\]<accompanimentIndicator\>

<composer(s)\> and <arranger(s)\> are entered in abbreviated form, with the abbreviation being checked against a list of abbreviations and expanded forms maintained in `names.json` (e.g. "JoRu" for "John Rutter"; "CtEcMr" for "Chris Tomlin, Ed Cash, Matt Redman"). If there are no composers, no text is required between the tilde ("~") and the underscore ("_"). if there are no arrangers, the underscore ("_") is not required and no text is required before the open square bracket ("\["). The choice of abbreviations is up to the individual.

<initialKeySignature\> is required. Specify the actual major or minor key where known, using `#` for sharp, `b` for flat and `m` for minor (e.g. `[C]`, `[F#]` or `[F#m]`). This allows fn2fm to write the forScore "Key" metadata.

If the actual key is not known, use the number of accidentals as a fallback: `[0]` for no sharps or flats, `[1#]` through `[7#]` for sharps, or `[1b]` through `[7b]` for flats. These counts produce no forScore key metadata and do not imply major or minor. They retain key-signature information in the filename, helping distinguish copies of the same song with different key signatures.

Empty brackets (`[]`) and brackets containing only spaces are rejected. Always supply either the actual key or the accidental count. Different keys can share the same accidental count, so use the actual key where known.

<accompanimentIndicator\> can be a plus ("+") for a tag that states "With Accompaniment", a hyphen/minus ("-") for a tag that states "Without Accompaniment", or left blank for no accompaniment tag.

## Name dictionary

The supplied `names.json` is a starter dictionary; customise it for your own scores.
fn2fm loads `names.json` from the current working directory (the folder from which
you run the command). You can keep it with the PDFs you work on and edit it there;
it does not need to be beside the executable.

When loading the dictionary, fn2fm warns if an abbreviation or name has leading
or trailing whitespace. Warnings identify the affected entry. Processing continues
with leading and trailing whitespace removed from names before they are used as
composer or arranger metadata. Abbreviations are not automatically corrected, and
the dictionary file is not rewritten. Spaces within names, such as `Dan Forrest`,
are preserved.

## Examples

In all these examples, the forScore Title will be the same as the filename, without the ".pdf" file extension

- O Magnum Mysterium ~ MoLa[2#].pdf
  - Composers: Morten Lauridsen
  - Arrangers:
  - Key:
  - Tags:
- Spirit Of The Season ~ GbAs_DaHa[Ab]+.pdf
  - Composers: Glen Ballard, Alan Silvestri
  - Arrangers: David Hamilton
  - Key: A♭
  - Tags: With Accompaniment
- African Noel ~ _AnTh[2b]+.pdf
  - Composers: 
  - Arrangers: André J Thomas
  - Key:
  - Tags: With Accompaniment
- Total Praise ~ RiSm[5b]-.pdf
  - Composers: Richard Smallwood
  - Arrangers:
  - Key:
  - Tags: Without Accompaniment
- Sing with Joy at Christmas (Stella Natalis) ~ KaJe[C]+.pdf
  - Composers: Karl Jenkins
  - Arrangers:
  - Key: C
  - Tags: With Accompaniment
- The Witness 16 The Victor ~ JaOc[0]+.pdf
  - Composers: Jamie Owens Collins
  - Arrangers:
  - Key:
  - Tags: With Accompaniment