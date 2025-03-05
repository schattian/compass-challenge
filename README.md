# compass challenge

This Go program takes an input filename of a contact list in CSV format and outputs the duplication scores in different formats to the STDOUT (see below for explanation)

## How to run

You can either:
- Run it directly with go (`go run ./... <INPUT_FILENAME>`)
- Run it with docker (`docker build -t compass .` and `docker run --rm -v "<INPUT_FILEPATH>:/data/input.csv" compass /data/input.csv`)

## Contents

- `labeled-output.csv`: contains the output produced by the app with labels for the different score ranges, as in the example
- `raw-output.csv`: contains the same output as `labeled-outputs`, but with the actual score number 
- `full-raw-output.csv`: contains all the scores, with no threshold (ie: carrying the whole n^2/2 comparisons, useless for deduplication purposes) 

## Other observations

- I avoided using 3rd party libraries on purpose (not because they're inherently bad, but IMO is better to show the solution in a more exhaustive way)
- Clocked timing working on the task: 3h40m