
## Host the formula

To try your formula via a browser, use the `runner` command line tool:

	cd web-embed-labs/
	./go/bin/runner ../formulas/ ./examples/test-probes/
	# Output should tell you which formula is hosted, like "some-name" that we used above

The second parameter, `../formulas/` is the parent directory of the page formula you created in the last step.

(Don't include "some-name" in the command. The runner will eventually be able to switch between formulas but for now it just loads the first formula in alphabetical order.)

The third parameter, `./examples/test-probes/`, points at a directory with JS for a few example test probes.

Point your browser at https://localhost/ (HTTPS and no port) and you should see the hosted page formula.

In the `runner` console output you should see any go template errors (usually from the page including template commands like `{{something}}`) or 404s. The formulator does its best but there is manual work involved with getting most page formulas cleaned up.

## Run test probes

Once you're looking at a hosted page formula (see above) you can run the test probes from the javascript console as if they were being called via Selenium.

In the javascript console, look at the `window.__welProbes` JS object to find which probes are loaded. There should be at least `dom-shape` and `exception` probes.

To run a probe:

	results = {}
	window.__welProbes["dom-shape"].probe(results)
	console.log(results)

The results object has test result key:value pairs.

Take a look in `web-embed-lab/examples/test-probes/` to see how probes are coded.



