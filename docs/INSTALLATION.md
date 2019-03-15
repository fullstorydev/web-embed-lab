# Web Embed Lab installation and initial build

The WEL will run on a UNIX-ish OS like macOS or Ubuntu. If you're on Windows you can use the Windows Linux Subsystem to host a linux distribution like Ubuntu.

## Prerequisites

You'll need to install the latest [git](https://git-scm.com/) and [Node](https://nodejs.org/en/download/) (to get npm) as well as [go](https://golang.org/doc/install) 1.11+.

You'll need `make` which on macOS usually means installing iCode. On a linux distro use the included package manager like yum or apt-get to install `make`.

## Clone and initial build

	git clone git@github.com:cowpaths/web-embed-lab.git
	cd web-embed-lab/fe/
	npm install
	# a lot of npm output here
	cd ..
	make
	# a lot of make output here
	# there should be several binaries in web-embed-lab/go/bin/

## First run of the colluder to generate certificates

During the target site capture process you'll need a few certificates so that the WEL can work with your browser. The first time you run the `colluder` tool it will generate these certificates.

	cd web-embed-lab/
	./go/bin/colluder
	# you should see a message about creating certificates

Use control-c to exit the colluder after it has created the certificates.

These certificates are set to expire a month from creation just so that they don't hang around forever. So, you'll need to delete them on occasion and then re-install them into your browser.

## Next steps

To develop new experiments (capturing page formulas, writing tests, etc) read [Developing experiments](EXPERIMENT_DEVELOPMENT.md).

To run existing experiments read [Running experiments](EXPERIMENT_RUNNING.md).
