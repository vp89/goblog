A simple CRUD blog app connecting to a Postgres database

My first attempt at writing Go code

### Todo

- [x] Save posts as Markdown, render to client as HTML
- [x] Add authentication to admin page
- [x] Improve index page layout
- [x] Improve post page layout
- [x] Don't require restart of the server to effect a template change
- [x] Add links to twitter/github/resume etc
- [x] Split out admin page between post index and add post
- [x] Add syntax highlighting to post code
- [ ] Refactor some helper functions to seperate files

### Usage

Needs a config.json file to connect to the Postgres:

{ "server":"SERVER_NAME_HERE", "Username":"USER_NAME_HERE", "Password":"PASSWORD_HERE", "Database":"DATABASE_NAME_HERE"}