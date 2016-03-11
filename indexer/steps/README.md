These steps comprise the pipeline used to get files indexed as well as to generate thumbnails for them.
Each step has a Start method which is called either by main or by the step it depends on (the one earlier 
in the pipeline). 
In addition, each step waits for its own queue to be empty AND for the preceding steps to be done.

In other words, `scanner` starts `checkindex` and then waits for `checkindex` to complete.

---

The `media` (add to ElasticSearch) pipeline is:

`scanner`:
- Scans the file systems for files to examine
- Passes the files to `checkindex`

`checkindex`:
- Checks ElasticSearch to determine if the file is already indexed or has changed since it was indexed.
- Passes the files to `getexif`
    - If it's an update, it passes to `generatethumbnail`
    - If it's in the index and not updated, it passes to `checkthumbnail`
    - If it's NOT in the index, it'll be checked after the `preparemedia` step
	
`getexif`:
- Invokes exiftool once for a directory, getting all exif info for each file.
- Passes the data to `preparemedia`

`preparemedia`:
- Parses the exif data, creating the media document
- Passes the data to `resolveplacename`
- Passes the data to `checkthumbnail`

`resolveplacename`:
- If the media has a location, do a reverse geocode to get the placename
- Passes the data to `indexmedia`

`indexmedia`:
- Adds/updates the media in the index, calling ElasticSearch
- < nothing else >


---

The `thumbnail` pipeline is:

`checkthumbnail`:
- Checks if the thumbnail exists
- Passes to `generatethumbnail` if not

`generatethumbnail`:
- Generates the thumbnail
- < nothing else >
