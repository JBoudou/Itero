This document provides a global view of the protocol between the front end and the middleware.

URL
===

URLs are partitioned as follows, based on the one-letter first segment.

 - **`/`** The root correspond to the front-end files (HTML, CSS and Javascript static files).
 - **`/s/`** Reserved for additional static files. May be abandonned in favor of Angular's assets.
 - **`/a/`** Public API, handled by the middleware. Details can be read in
   [main/main.go](../main/main.go).
 - **`/r/`** Virtual URL, handled by the front end. Details can be read in
   [app/src/app/app-routing.module.ts](../app/src/app/app-routing.module.ts). On the middleware,
   these requests are redirected to `/index.html`.

Details on the partition can be read in the function `server.Start` in file
[mid/server/server.go](../mid/server/server.go).

## Poll segment

To identify individual poll in URL, a dedicated poll segment is added to the URL. This segment is
always the last one. It consists of 9 alphanumeric characters. Poll segments are used both for
public API URLs and for virtual URLs. Examples: `/r/poll/iREs000cF` and `/a/info/count/iREs00cF`.


Method and query
================

The only accepted HTTP methods are GET and POST. GET must be used for all queries that do not need
any parameter (except for the poll identity which is encoded in the URL, see above). POST must be
used for all queries with parameters. In that case, the parameters are transmitted in the body of
the request, as a unique JSON structure. The JSON structure is described similarly in Go and
TypeScript, by an interface with a name ending with `Query`. Exactly one such interface is
associated with each URL. The file [app/src/app/api.ts](../app/src/app/api.ts) contains all those
interfaces for TypeScript. In the future, the TypeScript interfaces could be generated from the Go
ones.

Queries' parameters must never be transmitted in the query part of the URL.

POST queries must have an `Origin` header (or at least a `Referer` one). See method
`server.Request.CheckPOST` in file [mid/server/request.go](../mid/server/request.go).


Response
========

A status code between 200 and 299 indicates a success. In that case, the body is a JSON encoded
structure. The JSON structure is described similarly in Go and TypeScript, by an interface with a
name ending with `Answer`. Exactly one such interface is associated with each URL. The file
[app/src/app/api.ts](../app/src/app/api.ts) contains all those interfaces for TypeScript. In the
future, the TypeScript interfaces could be generated from the Go ones.

A status code of at least 300 indicates a failure. In that case, the body is a short unquoted string
describing the reason of the failure. Those strings are part of the API.

## Compression

When the user agent supports it, some responses may be compressed. To mitigate the BREACH exploit,
the middleware adds a random header of variable size. Nevertheless compression must be avoided for
responses containing arbitrary data from the request. As a rule of thumb, compression should be
allowed only for GET queries.


Session
=======

A session identifies a user across public API requests. Since the middleware is stateless (from the
point of view of the front end), sessions are stored only on the front end.

When a session starts (typically when the user successfully logged in), the middleware sends both a
session cookie and a session identifier. The session cookie is named `s` and is encrypted by a
private key (see `server.SessionKeys` configuration parameter). It contains the user identifier, the
session identifier and an expiration date. The session identifier is a 4 alphanumeric characters
string. It is send in the response's body.

For all subsequent request, the front end includes the session cookie and the session identifier in
each request. The session identifier is send in the HTTP header `X-CSRF`.

Sessions are created by the method server.Response.SendLoginAccepted in file
[mid/server/response.go](../mid/server/response.go). They are decoded by the method
server.Request.addSession in file [mid/server/request.go](../mid/server/request.go). On the front end,
sessions are handled by the classes [SessionService](../app/src/app/session/session.service.ts) and
[SessionInterceptor](../app/src/app/session/session.interceptor.ts).

## Unlogged users

Poll with Electorate field value 'All' can be accessed by anyone, even unlogged user. To identify
these voters and to allow them to change their vote on the next rounds, pseudo-users are created.
These pseudo-users are identified by a hash of their IP address. Moreover, a cookie is sent with the
user id and hash. This cookie is named `u` and is encrypted with the same private key as for session
cookies. When a request with this cookie is handled, the hash in the cookie is used in place of the
hash of the IP address.
