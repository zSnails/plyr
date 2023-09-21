# Music Player Backend - Plyr

Plyr is a straight forward music player backend written in Go. The only reason I'm
building this is for a college course so I'm probably never going to maintain it
after I'm done.

That's all the readme, I don't really care about this.

# TODO

- [X] Store song length

- [ ] Store song cover art

# Docker usage

## Build

```bash
docker build . -t plyr
```

## Run

```bash
docker run -it -p 8080:8080 plyr
```

## Add songs

```bash
# Copy songs to container
docker cp <song> <container_id>:/app/songs

# If container is stopped, start it
docker start <container_id>

# Add songs to database
docker attach <container_id>

# Now use add command for each song
>>> add
File> /app/songs/<song>
```