We discovered that in the case of boolean data types, golang barfs on string values, similar for string array, e.g. entrypoint, command, etc.

We are thinking we could just add a converter on top of the dockerclient structs to make them more tolerant. Other possiblities include ditching Docker compose alltogether or rewriting in Python.

Also, there is some strange behavior where a delete command was coming across the wire.

Also, make "build" work... Good luck with that suckas.
