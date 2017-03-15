# function-tracer
generic tracer for the whole stack

# Design
* Trace the target process, and store function invocation info in memory.
* Provide http interface to read them on demand.
* Do not store the trace in serialized form and durable to keep high performance.
* Correlate multiple function tracer data in the front-end if necessary
* Front-end is out of scope of this project

# TODO
* list target go process functions
* trace running go process
* trace running c process
* trace running php process