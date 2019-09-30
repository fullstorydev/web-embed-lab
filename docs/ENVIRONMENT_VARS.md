# Environment variables

There are a few common pieces of information needed when using the WEL command line tools so they're read as environment variables.

The easiest way to set and forget these is to create a "dotenv" file named `.env` in the directory where you run the commands.

Example:

	cp dotenv.example .env
	# Edit .env to enter your values

When a WEL command runs it looks in the current working directory for a `.env` file. The values it finds in that file will be used to set environment variables if they don't already exist when the command runs.

If you'd prefer to set the process env variables directly, see the [example bash script, setup-env.sh.example](../examples/setup-env.sh.example) for an example and instructions.

`BROWSERSTACK_USER` and `BROWSERSTACK_API_KEY` are used to authenticate with Browserstack. Once you have an account you can find your info in the "Automate" section of your [browserstack account page](https://www.browserstack.com/accounts/settings).   

`FRONT_END_DIST` holds the path to the `dist` directory created by `make fe`
