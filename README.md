# eclaire


![donuts-are-good's followers](https://img.shields.io/github/followers/donuts-are-good?&color=555&style=for-the-badge&label=followers) ![donuts-are-good's stars](https://img.shields.io/github/stars/donuts-are-good?affiliations=OWNER%2CCOLLABORATOR&color=555&style=for-the-badge) ![donuts-are-good's visitors](https://komarev.com/ghpvc/?username=donuts-are-good&color=555555&style=for-the-badge&label=visitors)

Eclaire is a lightning-fast static site webserver with automatic HTTPS written in Go. It is dead simple to use, fully portable, and automatically sets up HTTPS for your static sites in seconds. 

## why not nginx/apache?

NGINX and Apache are great, but they're general purpose webservers that are designed to handle a lot more responsibilities than just serving static content, and you still end up having to set up HTTPS after. 

Eclaire is your barely-there webserver, takes about 60 seconds to setup, and is designed from scratch to serve your [Bearclaw site](https://github.com/donuts-are-good/bearclaw) (or any static site!) with Let's Encrypt without interruption so your $5 droplet doesn't faint when your blog makes frontpage on HackerNews ;) 

By being able to focus on -just- serving static content, Eclaire has much less moving parts while accomplishing more in less time for static site deployments.

**Did you know?** Eclaire is both a sweet pastry and also the French word for 'lightning'

## usage

1. Download or build Eclaire, and put the `eclaire` binary wherever works best for you. 

2. Run `eclaire` to create your `www` folder, then place your sites in the www directory like this:

    - `./www/mycoolblog.com/`
    - `./www/whatever-subdomain.mycoolblog.com/`

3. That's it! Point your DNS at your server's IP like usual, and eclaire will start handling http and https requests!

**Note:** Eclaire is fully portable, and isn't putting files outside of its own folder where it was run from.

## eclaire with systemd

If you're on Linux, specifically a Systemd distribution, you can use Systemd to manage Eclaire with a "unit file" like this example here:

```
[Unit]
Description=Eclaire static webserver
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/eclaire
ExecStart=/path/to/eclaire/eclaire
Restart=always

[Install]
WantedBy=multi-user.target
```

## eclaire systemd cheatsheet

If you decide to let systemd manage eclaire for you, you can use these commands to manage your eclaire systemd service the same way you'd manage other webservers.

**Start eclaire** `sudo systemctl start eclaire `

**Stop eclaire** `sudo systemctl stop eclaire `

**Restart eclaire** `sudo systemctl restart eclaire `

**Start eclaire automatically at boot-time** `sudo systemctl enable eclaire `

**Do not start eclaire automatically at boot-time** `sudo systemctl disable eclaire `


## greetz

the Dozens, code-cartel, offtopic-gophers, the garrison, hedae, and the monster beverage company.

## license

this code uses the MIT license, not that anybody cares. If you don't know, then don't sweat it.

made with ☕ by 🍩 😋 donuts-are-good


## donate

If you would like to be an official energy drink sponsor of this project, you can contribute however you like.

**Bitcoin**: `bc1qg72tguntckez8qy2xy4rqvksfn3qwt2an8df2n`

**Monero**: `42eCCGcwz5veoys3Hx4kEDQB2BXBWimo9fk3djZWnQHSSfnyY2uSf5iL9BBJR5EnM7PeHRMFJD5BD6TRYqaTpGp2QnsQNgC` 

😆👏 Thanks
