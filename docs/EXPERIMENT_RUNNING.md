# Web Embed Lab: Running Experiments

Below you will find instructions for running Web Embed Lab (WEL) experiments. To understand experiments and how to create the read the [developing experiments](EXPERIMENT_DEVELOPMENT.md) document.

If you haven't already, follow the [installation instructions](INSTALLATION.md) to prepare the WEL tools.

The goal of running an experiment is to test how a specific version of an embedded script runs in a specific web page in a specific browser. If you are the author of a web site then you will want to test that the latest analytics script doesn't break or slow your pages. If you're the author of an embedded script then you will want to test that your new version doesn't break the web sites of your biggest users.

You will often want to run WEL experiments in your continuous integration system, but they can also be run locally in a development environment.

## Assemble your experiment

First you will need to assemble these things:
- page formulas
- test probes
- an embedded script
- experiment definition

You will find examples of all of these in web-embed-lab/examples/ and we will use those in the example commands below.

The command that runs the experiment is named `runner` and once you have successfully run `make` you will find it in web-embed-lab/go/bin/.

Here is an example of `runner` using the examples in the WEL repo:

	cd web-embed-lab/
	./go/bin/runner \
		./examples/page-formulas/ \
		./examples/test-probes/ \
		./examples/embed_scripts/no-op.js \
		./examples/experiments/hello-world.json

This command told the `runner` where to find the page formulas, test probes, experiment, and the embed script that is the focus of the tests. It also told the runner that we're using Browserstack and what authorization info to use.

The `runner` does the following:
- checks that all of the files are where they should be
- checks that the Browserstack environment variables are valid
- reads the experiment JSON
- runs tests against each page formula and browser combination defined in the experiment
- exits with 0 (all tests passed) or 1 (at least one test failed)


