# Search Images by Topic

A simple command line tool for processing images through the Clarifai API and then searching them by topic.


## Installation

```bash
$ go install github.com/gengel/image-tag-and-search
```


## Usage

### Build Index

Build a search index locally. Requires write access to disk and an API key.

To use the default list:

```bash
$ image-tag-and-search build -k "YOUR_API_KEY"
```

To use a custom list:

```bash
$ image-tag-and-search build -k "YOUR_API_KEY" -u "https://somedomain.com/files.txt"
```

"https://somedomain.com/files.txt" must be a newline-separated list of image urls.

### Search

Once the index is built, you can search it.

```bash
$ image-tag-and-search search "outdoors"
Search for flour
Found 5 matches for flour
https://c8.staticflickr.com/1/10/20788138_88c4458f8a_o.jpg
https://c7.staticflickr.com/3/2448/3694014992_3035f3d421_o.jpg
https://farm7.staticflickr.com/8167/7577887754_4751339cf2_o.jpg
https://farm8.staticflickr.com/5005/5356517508_b629917c6c_o.jpg
https://farm6.staticflickr.com/8484/8209261487_2956776817_o.jpg

```

## Future Work

* Add concurrency to initial build process
* Support multiple built indices simultaneously
* Support multiple models
* Options for downloading, viewing images
* Options for traversing index without searching by exact topic
