This document is a very detailed installation procedure from scratch on Windows.
It is meant for people wanting to have a dev server for contributing to Itero.

# Install Dependencies

Download and install the following open source softwares:

 - [MariaDB](https://mariadb.org/download/) version 10.3.28 or any more recent
   version. During the installation choose 'UTF8 as default character set' and
   provide a password for the root user.

 - [Go](https://golang.org/dl/) version at least 1.15.2.

 - Any recent version of [Node.js](https://nodejs.org/en/download/).

 - Any recent version of [Git](https://git-scm.com/downloads).
   Agreeing to all the default configuration should work.

# Fetch the sources

 - Create a [GitHub account](https://github.com/join) if you don't have
   one already.

 - Create an SSH key and add it to your GitHub account by following
   [this guide](https://docs.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent).
   (don't forget to follow the last step named "Add the SSH key to your GitHub
   account.").

 - Launch the "Git Bash" application then type `git clone git@github.com:JBoudou/Itero.git`.
   You should now have the sources of Itero in an Itero directory of your user
   directory.

# Install Angular

Launch "Windows PowerShell" then type the following:

```PowerShell
cd ~\Itero\app
npm install -g @angular/cli@10
npm update
cp .\karma.conf.js.stub .\karma.conf.js
```

# Create the database

Launch the program whose name looks like "Command Prompt (MariaDB 10.6 (x64))".
Then type
```Batchfile
mysql -u root -p
```
Enter now the password for the root user you have chosen during the
installation of MariaDB. Then type the following commands, replacing `passwd`
with the password of your choice (not one that you use for something else,
though).
```SQL
CREATE USER itero IDENTIFIED BY 'passwd';
CREATE DATABASE iterodb;
GRANT ALL PRIVILEGES ON iterodb.* TO itero WITH GRANT OPTION;
EXIT
```

Now the MariaDB user and database for Itero have been created. Let us populate
the database by typing what follows in the Command Prompt:
```Batchfile
cd C:\Users\Me\Itero
mysql -u itero -p iterodb
```
Of course, you'll have to replace `Me` in `C:\Users\Me\Itero` with your Windows
username. Then you have to give the password you have chosen for Itero.
Finally enter the following commands:
```MySQL
SOURCE sql\install.mysql
EXIT
```

# Config files

You first need to create some cryptographic keys. For that, launch "Git Bash"
and type the following commands.
```Shell
cd ~/Itero
openssl req -newkey rsa:4096 -x509 -sha256 -days 700 -nodes -out ssl.crt -keyout ssl.key
go build -o srvtool ./tools
./srvtools genskey
```
The last command should display a character string starting and ending with
square brackets. You will need this string.

Create the file config.json in the Itero directory, with a content like the
following, except that on the line with "SessionKeys" you copy the string from
the previous step.
```JSON
{
  "database": {
    "DSN": "itero:passwd@tcp(localhost)/iterodb?loc=Local"
  },
  "server" : {
    "Address": "localhost:8443",
    "SessionKeys": ["what_was_given_by_srvtool_genskey=","same=="],
    "CertFile": "ssl.crt",
    "KeyFile": "ssl.key"
  }
}
```

# Tests

Launch "Windows PowerShell" and type the following.
```PowerShell
cd ~\Itero
go test -cover ./...
cd app
ng test --watch=false
```
Of course, all the tests must pass successfully. You may still have errors
reading `Error 1615: Prepared statement needs to be re-prepared`. These errors
are caused by a bug in MariaDB. To workaround this bug, you need to add
`&autoReprepare=1` to the `DSN` parameter in config.json, such that the line
looks like:
```JSON
    "DSN": "itero:passwd@tcp(localhost)/iterodb?loc=Local&autoReprepare=1"
```

# Build and Try

Launch "Windows PowerShell" and type the following.
```PowerShell
cd ~\Itero\app
ng build
cd ..
go build -o srv.exe ./main
.\srv.exe
```
Itero is now running!

To try it, open any web browser at the address https://localhost:8443/. You will
usually see an alert page saying that the page you want to access is not secure
or even dangerous. This is only because we use self-signed SSL certificate. In
fact, there is absolutly no risk at all, because it's your own site, on your
own computer. To escape that warning and access Itero, first click on a button
called "Advance setting" or something similar, and then click on another button
saying something like "Continue at your own risk".
