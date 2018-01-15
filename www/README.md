# www
This contains all the resources and code related to the moebot website

## Building and developing
* Install `github.com/jteeuwen/go-bindata` (it should've been installed already during setup)
* If you ever change any files in the `/static` directory, re-run `go-bindata static/...`
* `docker-compose up --build` will compile any new changes and bring up the whole project
* Navigate to `localhost:8080` to view the website

## Notes
Currently it's pretty annoying to create or update static files, you've got to modify them,
take down docker containers, `go-bindata`, and bring it back up.
I'm going to figure out a solution to this soon. 