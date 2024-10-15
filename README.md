### Simpl Lang

## This repository is a work in progress.

# Declaring a variable

```
var v1 int
var v2 int = 2
var v3 string = "Hello"
var v4 float = 2.14
var v5 array[string] = ["a", "b", "c"]
var v6 array[array[int]] = [[1,2,3], [3,4,5]]
```

# Declaring a constant

```
const v1 int = 1
const v2 string = "Hi!"
```

# Reading from stdin

```
read v1
```

# Writing to stdout

```
write "This is the value of the v1 variable: ", v1, " and that's it" 
write v2
```

# if statement

```
if v1 == 2 {
    write "v1 is 2"
}

if v2 >= 2 {
    write "v2 is greater than 2"
} else {
    write "v2 is less than 2"
}

if v3 == v4 {
    write "v3 is greater than v4"
} else if v3 < v4 {
    write "v3 is less than v4"
} else if v3 > v4 {
    write "v3 is greater than v4"
}
```

# for loop

```
// Should print the values 0, 2, 4

from i from 0 to 5 step 2 {
    write i, " "
}
```

# while loop

```
// Should print 0 1 2 3 4 5

var i int = 0
while i <= 5 {
   write i, " " 
}
```
