myq-tools
=========

Tools for monitoring MySQL (successor to myq_gadgets)

Project Status
---------------
This package is under development and not in a prod release yet.

Tools
-----
Currently there is a single tool, 'myq_status'.  More tools may be added in the future.

* **myq_status**: Iostat-like views of MySQL SHOW GLOBAL STATUS variables.  Use '-help' to get more detail on available views.

Binaries
--------
Binaries are available in the Releases tab here in Github. 

Running development/latest version
----------------------------------
1. Download and install golang.
2. Clone this repo to your GO home
3. Execute 'go run myq_status.go'

