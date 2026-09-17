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

## fn2fm Score Naming Convention

The naming convention used by fn2fm has been developed to support the needs of our choir [Cambrensis](https://www.cambrensis.org.uk/), uses only [7-bit ASCII printable characters](https://www.ascii-code.com/characters/printable-characters) constrained by the characters to be avoided listed above, and supports the following metadata:
- Score Title (the same as the PDF filename, but without the ".pdf" extension)
- Composer(s)
- Arranger(s) (used by repurposing "Genre(s)")
- Initial key signature
- Whether or not the score includes an accompaniment (this is visible in the forScore "Tag(s)")

The filename is constructed from two parts separated by a semicolon (";"):
- The score title
- The score metadata

To minimise uncertainty, the filename must contain exactly one semicolon. If there is no semicolon present, there is no metadata and therefore nothing for fn2fm to do. If there are two or more, fn2fm will not know which separates the score title from the metadata and the filename will be rejected.

The score metadata is formatted as follows and, apart from the limitation of no trailing spaces in the filename, each can have leading and trailing spaces if they aid readability (we just use a single space after the semicolon):
- <composer(s)\>_<arranger(s)\>\[<initialKeySignature\>\]<accompanimentIndicator\>

<composer(s)\> and <arranger(s)\> are entered in abbreviated form, with the abbreviation being checked against a list of abbreviations and expanded forms maintained in ... (e.g. "JoRu" for "John Rutter"; "CtEcMr" for "Chris Tomlin, Ed Cash, Matt Redman"). If there are no composers, no text is required between the semicolon (";") and the underscore ("_"). if there are no arrangers, the underscore ("_") is not required and no text is required before the open square bracket ("\["). The choice of abbreviations is up to the individual.

<initialKeySignature\> can be entered as either the number of accidentals (sharps or flats), or the major or minor key. The number sign/hash ("#") is used for the sharp symbol and lowercase "B" ("b") for the flat symbol (e.g. "0" or "C" for C major; "1#" for one sharp; "2b" for two flats; "F#" for F sharp major; "F#m" for F sharp minor). The forScore "Key" metadata can only be updated if the major or minor key signature is provided. If no key signature is provided, empty square brackets ("[]") should still be included in the filename.

<accompanimentIndicator\> can be a plus ("+") for a tag that states "With Accompaniment", a hyphen/minus ("-") for a tag that states "Without Accompaniment", or left blank for no accompaniment tag.

## Examples

In all these examples, the forScore Title will be the same as the filename, without the ".pdf" file extension

- O Magnum Mysterium; MoLa[2#].pdf
  - Composers: Morten Lauridsen
  - Arrangers:
  - Key:
  - Tags:
- Spirit Of The Season; GbAs_DaHa[Ab]+.pdf
  - Composers: Glen Ballard, Alan Silvestri
  - Arrangers: David Hamilton
  - Key: A♭
  - Tags: With Accompaniment
- African Noel; _AnTh[2b]+.pdf
  - Composers: 
  - Arrangers: André J Thomas
  - Key:
  - Tags: With Accompaniment
- Total Praise; RiSm[5b]-.pdf
  - Composers: Richard Smallwood
  - Arrangers:
  - Key:
  - Tags: Without Accompaniment
- Sing with Joy at Christmas (Stella Natalis); KaJe[C]+.pdf
  - Composers: Karl Jenkins
  - Arrangers:
  - Key: C
  - Tags: With Accompaniment
- The Witness 16 The Victor; JaOc[0]+.pdf
  - Composers: Jamie Owens Collins
  - Arrangers:
  - Key:
  - Tags: With Accompaniment