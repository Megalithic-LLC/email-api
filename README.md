# On-Prem Email API server

API Server for the on-prem cloud administration portal.

## Building

```sh
$ make
```

## Running

```sh
$ make
$ cd $GOPATH/bin
$ ./on-prem-email-api
2019/09/22 01:56:10 INFO Attached to MySQL
2019/09/22 01:56:10 INFO Completed migrations
2019/09/22 01:56:10 INFO Attached to Redis at localhost:6379
2019/09/22 01:56:10 INFO Listening for http on port 3000
2019/09/22 01:56:10 INFO On-Prem Email API started
```

Then clone, build, and run the [On-Prem Email Server](git@github.com:on-prem-net/emaild.git):

```sh
$ git clone git@github.com:on-prem-net/emaild.git
...
$ make
$ cd $GOPATH/bin
$ ./on-prem-emaild
2019/09/22 02:05:25 INFO Opened database /.../on-prem-emaild.db
2019/09/22 02:05:25 DEBUG Ensuring indexes
2019/09/22 02:05:25 INFO Listening for IMAP4rev1 on :8143
2019/09/22 02:05:25 DEBUG Connected to ws://localhost:3000/v1/agentStream
2019/09/22 02:05:25 INFO On-Prem Email Server started
2019/09/22 02:05:25 INFO Node id is bm3jict5jj84bcnaud40
```

Finally, clone and run the [On-Prem Email Console](git@github.com:on-prem-net/on-prem-email-console.git), and login with the default credentials of `admin` and `password`:

```sh
$ git clone git@github.com:on-prem-net/email-console.git
$ cd email-console
$ npm install
$ ember serve --proxy=http://localhost:3000
Proxying to http://localhost:3000

Build successful (8113ms) – Serving on http://localhost:4200/
```
