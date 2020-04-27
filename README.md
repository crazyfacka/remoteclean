# remoteclean
Cleanup files when a given remote mount is low on space. It deletes the oldest files in the whitelisted folders.

## Requirements

You'll need:

* Go >= 1.14
* SSH agent running on the machine, with support for public/private key authentication

## Build & configure

### Build and copy sample files

```bash
go build
cp sample.remoteclean .remoteclean
```

### Edit configuration

Edit the configuration file `.remoteclean` which is already populated with some sample configuration parameters.

**mount** mount point to check for available space<br/>
**space_threshold** floating point number representing the GB on which to trigger the deletion of contents<br/>
**remote** information regarding the remote host<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*host* - hostname or IP address of the machine<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*port* - port to where the SSH agent is listening<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*user* - user with which the application should login<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*key* - location of the private key to use to SSH<br/>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*dirs* - array representing the directories from which content should be deleted

### Add keys

You need to add the private keys used to SSH into both the seedbox and the player machine. You also need to have them already added to your `~/.ssh/known_hosts` file. So make sure that you've at least SSH'd to those machines once through your terminal (easier way to add them to the file).

## Execute

```bash
$ ./remoteclean
```