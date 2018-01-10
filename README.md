# dp-filter-down-load-test

Run 10 parallel filter download jobs & capture their individual times & the overall time.

## Getting started

Add the instanceID of the instance you wish you filters to run against in `config.yml`.

Create/modify a filter json file under `/filters` - this is a filter job which will be submitted to the filter API. 
If you added a new file then add the filename `filters` field in `config.yml`

Compile
```.bash
go build -o loadtest
```

Run it
```.bash
HUMAN_LOG=1 ./loadtest
```

