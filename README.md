# Web Embed Lab

The Web Embed Lab (WEL) provides tools for testing that a specific version of an embedded script does not negatively effect the web page in which it is embedded. Many web sites embed dozens of third party scripts and as a result are often slow to download and slow to render.

This tool helps embedded script writers and web site developers test embedded scripts to identify when there is a problem.

Examples of embedded scripts are:
- [FullStory](https://www.fullstory.com/)'s session capture and playback
- Third party analytics like Google Analytics 
- Social media integrations like Twitter or Facebook's
- Continuous integration / testing status scripts like [Coveralls](https://coveralls.io/)

The WEL works with "experiments" which are bundles of files that hold all of the information required to test specific pages on specific browsers.

Unlike other tools, the WEL is able to capture a remote web site into a "page formula" that never changes. That way tests can be run with consistent results and tests don't add load to the remote web site.

The overall workflow of the WEL is to capture remote web sites, write tests, and then use [WebDriver](https://www.w3.org/TR/webdriver1/) and [Selenium](https://docs.seleniumhq.org/) to run those tests on those captured sites in real browsers.

People using the WEL generally work in two separate phases: developing experiments and then later running experiments in their continuous integration system.

These documents cover those topics:
- [Developing experiments](./docs/EXPERIMENT_DEVELOPMENT.md)
- [Running experiments](./docs/EXPERIMENT_RUNNING.md)

If this is your first time setting up the WEL then read the [Installation guide](./docs/INSTALLATION.md).

![Overview diagram](https://cowpaths.github.io/web-embed-lab/images/wel-components-002.png)

