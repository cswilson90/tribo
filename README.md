# Tribo - A static blog generator

Named after the [Triboelectric effect](https://en.wikipedia.org/wiki/Triboelectric_effect) which is a cause of static electricity.

I created this for my own personal use but feel free to use it if you want.

## Quick Start

To quickly get the example blog running run the following commands:

```
$ git clone https://github.com/cswilson90/tribo.git
$ cd tribo/
$ make
$ cd example
$ ../output/tribo --outputDir /srv/blog
```

This will build the example blog and install it in `/srv/blog`.
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

## References

The example given makes use of [list.js](https://listjs.com/).
