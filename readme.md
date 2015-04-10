A simple CRUD blog app connecting to a Postgres database

My first attempt at writing Go code

### Todo

- [ ] Save posts as Markdown, render to client as HTML
- [ ] Add authentication to admin page
- [ ] Improve index page layout
- [ ] Improve post page layout
- [ ] Don't require restart of the server to effect a template change
- [ ] Add links to twitter/github/resume etc
- [ ] Split out admin page between post index and add post
- [ ] Add syntax highlighting to post code
- [ ] Show success/failure on admin page for any insert/edit/delete operations

### Usage

Needs a config.json file to connect to the Postgres:

{ "server":"SERVER_NAME_HERE", "Username":"USER_NAME_HERE", "Password":"PASSWORD_HERE", "Database":"DATABASE_NAME_HERE"}