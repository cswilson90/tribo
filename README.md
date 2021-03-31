# Tribo - A static blog generator

Named after the [Triboelectric effect](https://en.wikipedia.org/wiki/Triboelectric_effect) which is a cause of static electricity.

I created this for my own personal use but feel free to use it if you want.

## Quick Start

First you need to install the Tribo executable:

```
go get github.com/cswilson90/tribo/cmd/tribo
```

This will install the tribo executable in `$GOPATH`. It's recommended you add
the location of `$GOPATH` to `$PATH` so the program can be executed anywhere.

To quickly get the example blog running run the following commands:

```
$ git clone https://github.com/cswilson90/tribo.git
$ cd tribo/example
$ tribo
```

This will build the example blog and install it in `/srv/blog`.

If you have your own blog directory already set up you should run the tribo
command from that directory.

To view the blog you will then need to install and run a webserver.
Below is an example nginx config to serve the site.

```
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    root /srv/blog;

    index index.html;

    server_name _;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

You can use any webserver. You just need to configure it to serve static files out of `/srv/blog`.

You can then view the example blog by visiting `http://127.0.0.1/` in a browser on the machine
running the webserver.

## Blog Directory Layout

The [example](example/) directory gives an example layout of a blog. This uses the default name
for each directory but you can change the directory names as described in the
[program configuration](#program-configuration) section below.

```
example/
+--posts/
|  +--2021/
|     +--01/
|     |  +--content.md
|     |  +--metadata.yaml
|     +--03/
|        +--image-post/
|        |  +--resources/
|        |  |  +--cat.jpg
|        |  +--content.md
|        |  +--metadata.yaml
|        +--post-2
|           +--content.md
|           +--metadata.yaml
|
+--static/
|  +--blog.css
|  +--blog.js
|
+--templates/
|  +--includes/
|  |  +--header.html.tmpl
|  |  +--footer.html.tmpl
|  +--post.html.tmpl
|  +--post_list.html.tmpl
|
+--.tribo.yaml
```

### Post Files

`posts/` is where you should but all the content and config for your individual blog posts.

Tribo will recursively walk the `posts/` directory looking for posts. A post is any directory
that contains both a `content.md` and `metadata.yaml` file. If a directory contains both files
the program will look no further down the directory tree so post directories can't themselves
contain sub-directories which are posts.

The example orders the posts by year and month but you can order the directories in whatever
way you want. However, no matter how you order the input directories the output will be grouped
by into directories by year and month of publication.

As well as year and month directories each post is also given it's own directory based on it's
title. In the example the post at `posts/2021/01/` is available at
`http://127.0.0.1/2021/01/my-first-post/`.

A post directory can contain the following:

* `content.md` (required) - a markdown file containing the content of the blog post. The first thing
  in the content file should be a heading with the title of the post. The content of the heading is
  extracted and used as the title. The first paragraph of the content is extracted and used as a
  preview on the posts list page.
* `metadata.[yaml|json]` (required) - a YAML or JSON file containing metadata for the the blog post,
  see the [post metadata section](#post-metadata) for information on the data that can be provided.
* `resources/` (optional) - a directory containing static resources used in the post e.g. images.
  These will be copied to the root of the output blog post directory e.g. in the example
  `image-post` uses an image at `http://127.0.0.1/2021/03/a-post-with-an-image/cat.jpg`. This is
  linked to in `content.md` using a relative link e.g. `![Cat Image](cat.jpg)`

### Static Files

`static/` is where you should put any static resources that will be used throughout the site
such as javascript, CSS and images.

The resources will be copied to the root directory of the output so will be available
at e.g. `http://127.0.0.1/blog.js` for `blog.js` in the example

### Template Files

`templates/` is where you should put templates for generating the static pages. See the
[writing your own templates](#writing-your-own-templates) section for more information about
customising the templates.

### Config File

This is a YAML config file that can be used to set Tribo run time options. See the
[program configuration](#program-configuration) section for information on what can be
configured in the file. By default running `tribo` in the directory with no arguments
will load config from this file.

The YAML should be a single hash where each key is a config option e.g.

```
---
outputDir: /srv/blog
blogName:  "My Blog"
```

## Post Metadata

The metadata for a post should be given in a file called `metadata.yaml` or `metadata.json`
in the post's directory.

The following options can be given:

| Option      | Required | Description                                                                                                        |
|-------------|----------|--------------------------------------------------------------------------------------------------------------------|
| publishdate | Yes      | The date of publishing of the post. This is used to generate the link for the post. Should be in `YYYY-MM-DD` format. Posts with a publish date in the future won't be added to the output. |
| tags        | No       | A list of tags to attach to the blog post.                                                                         |
| linkname    | No       | The name used as the last part of the link to the post. If not given a name will be generated from the post title. |

An example of the contents of a metadata YAML file:

```
---
publishdate: "2021-03-10"
tags:
  - fun
  - aboutme
```

## Program Configuration

Tribo has several configuration options that can be set either with a command line argument or in
a [YAML config file](#config-file).

By default the program will try to read config from a file called `.tribo.yaml` in the working
directory (generally the directory the program is executed in). You can specify a different config
file by supplying the `-configFile` argument when running the `tribo` executable. If you don't
specify a config file when executing the program and no `.tribo.yaml` file exists the default
value will be used for all options.

The following can be configured:

| Name        | Default Value  | Description                                                                                                                                                                                                      |
|-------------|----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| baseURLPath |                | The base URL path of the blog if it isn't the root of the site e.g. if you wanted your blog at `http://127.0.0.1/blog/` you should set this to `/blog`                                                         |
| blogName    | My Blog        | The name of the blog. This is passed to the templates when generating the site.                                                                                                                                  |
| blogDescription | My musings about the world | The description of the blog. This is passed to the templates when generating the site and is used as the description of the RSS feed.                                                            |
| rssLinkUrl  | `http://127.0.0.1` | The link used in the RSS feed to link to posts. For normal navigation the links are relative but the RSS feed needs an absolute URL. This will be joined with the `baseURLPath` to build links to add to the RSS feed. |
| outputDir   | `blog`         | The directory to output the static blog files to. Default is `blog/` in the working directory.                                                                                                                   |
| postsDir    | `posts`        | The directory where the raw content of the blog posts are saved. Default is `posts/` in the working directory.                                                                                                   |
| staticDir   | `static`       | The directory where static resources for the entire blog are saved. The contents of the directory is copied into the output directory to be served by the server. Default is `static/` in the working directory. |
| templateDir | `templates`    | The directory which stores the templates used to generate the pages of the blog. Default is `templates/` in the working directory.                                                                               |
| parallelism | Number of CPUs | The max number of blog posts generated in parallel. Defaults to the number of CPUs available on the machine.                                                                                                     |
| futurePosts | `false`        | Whether to publish posts with a publish date in the future. Values in YAML should be `true`/`false`, the flag can be given with no arguments on the command line to enable.                                      |
| noOutputCleanup | `false`    | By default Tribo will delete any directories from the output directory that it thinks are from posts which no longer exist or have been moved due to a title or published date change. You can set this option to `true` to stop this behaviour if it is causing problems. |

## Writing Your Own Templates

Templates use golang's `html/template` [package](https://golang.org/pkg/html/template/).

There are two main template files in the templates directory:

* `post.html.tmpl` - used to generate the page for a single post
* `post_list.html.tmpl` - used to generate the list of posts that is used as the main page of
the blog

The template folder also contains a `includes/` directory in which you can put templates which
are included in the two main files. In the example this is just the header and footer but more
can be added if required.

A postPageData object is passed in as the input to `post.html.tmpl` and a postListPageData object
is passed to `post_list.html.tmpl`. The structure of the objects is as follows:

```
postPageData {
    Common: commonData, // Data common to all pages
    Post:   postData,   // Data for the post to be displayed on the page
}

postListPageData {
    Common:  commonData,   // Data common to all pages
    Posts:   [ postData ], // A list of data for each post in the list (sorted by publish date)
    AllTags: [ string ],   // A list of all tags from all posts (ordered alphabetically)
}

commonData {
    BaseURLPath:    string, // The base path of the blog on the server
    BlogName:       string, // The global name of the blog
    BlogDescription string, // The global description of the blog
    PageTitle:      string, // A title for the page to be used as the HTML title
    CurrentYear:    string, // The current year as a string (for use in copyright notice)
}

postData {
    Title:       string,        // The title of the blog post
    Content:     template.HTML, // The HTML content of the post (in a format compatible with the `html/template` package)
    Preview:     template.HTML, // The HTML content of the preview of the post
    PublishDate: string         // The publish date of the blog post in "01 Jan 2000" format
    Url:         string         // The direct URL link for the post
    Tags:        [ string ]     // A list of tags attached to the post
}
```

## References

The example given makes use of [list.js](https://listjs.com/).
