# Web Embed Lab: Developing Experiments

If you haven't already, follow the [installation instructions](INSTALLATION.md) to prepare the Web Embed Lab (WEL) tools for use.

## Workflow summary

When developing experiments you will:
- Capture target web sites
- Create reusable page formulas
- Write test probes
- Choose which page formulas, tests, and browsers make up the overall experiment

### Capture target web sites

A "target web site" is a live site that you'd like to ensure works with an embedded script.

For example, if you run your own web site then you will want to capture it. If you're the creator of an embedded script then you will want to capture the sites of your biggest users.

### Create reusable page formulas

Once you've captured a target site then you'll want to create a "page formula" which is sort of a "freeze-dried" version of the target site that can be tested by the WEL without directly connecting to the target site.

Page formulas hold all of the HTML, CSS, Javascript, and image files of the target site and all of the information needed to load those assets into a test page.

### Write test probes

The WEL comes with a number of "test probes" that you can use to test that a page formula running in a browser has not been damaged in any way by an embedded script. Test probes can look at any aspect of the page, including how many unhandled exceptions have been thrown or whether a certain DOM structure exists.

You will often write your own test probes to make sure that specific aspects of the embedded script aren't a problem.

### Create an experiment

An experiment pulls together all of the page formulas, test probes, and types of browsers that you'd like to test. That information is held in a JSON file that is then handed to the WEL runner (as explained in [Running experiments](EXPERIMENT_RUNNING.md)).

## How to capture a target web site and formulate a page formula

While you can capture browser sessions using your local browser, it involves loading a certificate into your browser and setting it to use an HTTP proxy. It is much easier to use the `auto-formulate` command which uses Browserstack to configure a browser, capture sessions, and then create page formulas.

If you'd like to capture locally, read [capturing locally](CAPTURE_LOCALLY.md).

The `auto-formulate` command reads a JSON configuration that defines which browsers to spin up, which sites to capture, and then how to translate (or formulate) those captures into page formulas.

Look in web-embed-lab/examples/auto-formulate/ for examples of those configuration files. The easiest way to get started is to copy and modify one of those files.

To automatically capture web sessions and formulate page formulas:

	cd web-embed-lab/
	source set-env.sh
	./go/bin/auto-formulate \
		./examples/auto-formulate/hello-world-formulate.json \	# A config file
		../pf/	# The destination dir for new page formulas


## Refining a page formula

Once you have a page formula you'll probably need to edit it a bit to make it a stable page for test probes.

The `runner` command used when running experiments (explained in [Experiment running](EXPERIMENT_RUNNING.md)) has a developer mode that enables you to host a specific page formula and look at it with your browser.

Assuming that you have run `make` to build the `runner` command (explained in [Installation](Installation.md)) you can put the `runner` into development mode like this:

	cd web-embed-lab/
	./go/bin/runner \
		./examples/page-formulas/ \	# a directory holding page formula sub-directories
		./examples/test-probes/		# a directory holding test probe sub-directories

You should see a message that the `runner` is hosting one of your formulas and listening on port 9090.

You can point your web browser at http://127.0.0.1:9090/ to see the currently hosted formula.

The `runner` only hosts one page formula at a time but it knows about all of the formulas in the directory of page formulas that you passed it above.

You can list the available page formulas and learn which formula is currently hosted by GETing this URL:

	curl http://127.0.0.1:9090/__wel_control

The return value will list the formulas and which formula is currently hosted:

	{
		"formulas":[
			"vanilla-site",
			"hello-world"
		],
		"current-formula":"vanilla-site"
	}

You can change which page formula `runner` is hosting by PUTing to the same URL:

	curl http://127.0.0.1:9090/__wel_control -X PUT --data "{\"current-formula\":\"PAGE_FORMULA_NAME\"}"

If you want to test a page formula with an embed script then use the embed mode of the `runner`:

	cd web-embed-lab/
	./go/bin/runner \
		./examples/page-formulas/ \	# a directory holding page formula sub-directories
		./examples/test-probes/	\	# a directory holding test probe sub-directories
		./examples/embed_scripts/no-op.js # the embed script 

