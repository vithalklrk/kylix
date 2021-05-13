# kylix

Kylix is a utility to watch ipTV channels on your monitor. It's written in Go language

The binaries (.deb file) are available only for Debian-based systems like Ubuntu, Mint etc. Just download the release and use the command 

`  dpkg -i kylix_1.0.deb`
  
For other systems, it's simple to build and install from the sources, as described below

- First, you will need [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [golang](https://golang.org/dl/) installed on your system
- Next, clone the kylix repository with the command `git clone https://github.com/vithalklrk/kylix`
- Then `cd` to the cloned directory and run `go build`. If the command is successful, then the executable `kylix` should be created
- Next, run the command `sudo sh ./install.sh`. You will be asked for your password and then the script copies the files to your system
- Now you can execute `kylix` from the command line or through your UI
- To uninstall, run `sudo sh ./uninstall.sh`
