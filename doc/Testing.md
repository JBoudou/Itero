Testing the middleware
======================

To test all packages of the application simply launch `go test -cover ./...` in the root directory.

Testing the front end
=====================

Currently, there are only unit tests using Karma. These tests depends on the configuration file
`app/karma.conf.js`. That file is not part of the git repository in order to allow the tests to be
customize for each installation. For instance, the tests could be run without any browser on the
production machine. The file [app/karma.conf.js.stub](../app/karma.conf.js.stub) provides a working
starting point for the configuration file.

Once the configuration file has been created, launch the test by running `ng test` in
[app/](../app) directory.
