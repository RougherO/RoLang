# RoLang

RoLang is a tree-walk interpreter that is built using pure Go. It is an educational endeavour aimed at understanding the working of interpreters and parsing techniques like Pratt-Parsing, and is inspired from Thorsten Ball's monkey interpreter, but it also deviates in places where it did not suit my tastes.

It has an extremely easy language syntax, making it suitable for beginners new to programming.

## Updates
1. After the 0.4 update, lots of refactoring went in, although it passed the previous tests, I do think the tests are weak at this moment, so language could be buggy in some places, especially nested functions and closures and probably loops, feel free to break it any way possible and report new issues :D

## At a Glance

- ### Dynamic typing

    RoLang is strongly dynamically typed, like python. You can assign any type of value to any variable but only allows operations which are permissible for the type

    ```
    |> io.println(1 + true);

    repl:1:1: 
    repl:1:3: 
    repl:1:7: addition not supported for bool
    ```
- ### Variables

    You can create new variables using the `let` keyword
    ```
    |> let x = 1;

    |> let y = 2;

    |> io.println(x + y);
    3
    ```

    However redefinition or shadowing of variables in same scope is not allowed
    ```
    |> let x = 1;

    |> let x = 2;

    repl:1:1: variable x already exists in current scope
    ```
    You can create a new scope and define the variable there. Defining a new scope is done using `{` and `}`.
    ```
    |> let x = 1;

    |> { let x = 2; io.println(x); }
    2

    |> io.println(x);
    1
    ```
    
    RoLang has the following data types
    
    | type      | example                 |
    |-----------|-------------------------|
    | int       | 1, 10                   |
    | float     | 1.5, 2.23               |
    | string    | "hello", "bob"          |
    | bool      | true, false             |
    | functions | fn () { return x + y; } |
    | arrays    | [1, 2.2, "hey"]         |
    | maps      | {"a": 1, "b": 2}        |

    > [!NOTE]
    > Maps can have only strings, ints, floats and bools as key type

- ### Functions
    
    Declaring new functions is done using `fn` keyword
    ```
    |> fn greet(name) { return "Hello " + name; }

    |> io.println(greet("me"));
    Hello me
    ```

    Funcions are first class objects, which means you can even pass them around as values or create one using function literal syntax like below
    ```
    |> let add = fn (x, y) { return x + y; };

    |> io.println(add(1, add(2, 3)));
    6
    ```

    It also supports closures
    ```
    |> let add = fn(x) { return fn(y) { return x + y; }};

    |> let add2 = add(2);

    |> io.println(add2(3));
    5
    ```
- ### Operators

    RoLang supports quite a bit of operators, not as many as something like C++ though. Below are some of them:

    | operators | data types supported      |
    | --------- | ------------------------- |
    | +         | int, float, string,       |
    |           | arrays, maps              |
    | -         | int, float                |
    | *         | int, float                |
    | /         | int, float                |
    | <, >      | int, float, string        |
    | <=, >=    | int, float, string        |
    | == , !=   | int, float, string, bool, |
    |           | arrays, maps              |

    Addition for strings concatenates them, for example 
    ```
    "Hello" + "World" => "HelloWorld"
    ```
    same thing for arrays as well
    ```
    [1, 2, 3] + [3, 4] => [1, 2, 3, 3, 4]
    ```
    For maps however, addition of them does a union of key-value pairs and if two maps have same key then only one key exists in the final map with the value of second map as the final value
    ```
    |> let collection1 = {"a": 1, "b": 2};
    
    |> let collection2 = {"a": 2, "c": 3};
    
    |> let collection = collection1 + collection2;

    |> io.println(collection);
    {"a": 2, "b": 2, "c: 3};
    ```
- ### Top-level Return Statements
    
    Return statements in general are used to return values from function calls. However using return statements at global level, i.e., outside any function returns value as a process and exits
    ```
    |> return 0;
    ```
    exits with an exit status 0 while the one below
    ```
    |> return 2;
    ```
    returns with an exit status 2,

    Returning with no value also exits with exit code 0 indicating success
    ```
    |> return;
    ```
    However returning anything other than integers is an error
    ```
    |> return 1.2; // OR return "hello"
    ```

- ### Control-Flow

    There's very primitive support for control-flow. Right now, it only contains a single loop statement with an optional condition
    
    ```
    |> loop { io.println("yes"); }
    ```
    
    This will infinitely keep printing 'yes`, to put a condition check you can put an additional condition expression which is evaluated before every iteration
    
    ```
    |> let x = 2;

    |> loop x > 0 { io.println(x); x = x - 1; }
    2
    1   
    ```
    It also has if else statements.    
    ```
    |> let x = true;
    
    |> if x { io.println("yes"); } else { io.println("no"); }
    yes
    
    ```
    If-else statements can be chained as well
    ```
    |> if x == 2 { io.println("have 2"); } else if x == 3 { io.println("have 3"); }
    ```

    It doesn't have support for `break` and `continue` yet

- ### Standard Library

    Perhaps the best feature of RoLang is its highly extensible and customisable standard library. It comprises of multiple modules that are baked into the language. Why extensible? Because it is very easy to write your own standard library module or function for an existing module and hook it up with the existing code, with very minimal changes. In fact we will look into an example soon, here are the current modules present in the standard library. One does not need to import them to use them, they are pre-imported automatically.

    Syntax for calling module functions is like this
    ```
    |> <moduleName>.<functionName>(<arguments>...);
    ```
    - `io`: This module deals with all I/O operations.
        
        - `readln`: Can read a line from stdin. It reads all characters until it encounters a newline, and returns it but discards the newline. It does not take any parameters
        
        - `print`: Can take a variable number of arguments and print it to the stdout and flushes after every call. It prints each of the items without any separation in between
        
        - `println`: Same as print except it additionally prints a newline (`\n`)
        
        Example
        ```
        |> let line = io.readln();
        Hello World
        
        |> io.println(line);
        Hello World

        |> io.println("Hi there " + "Yoru!");
        Hi there Yoru!
        ```

    - `strings`: This module deals with string operations
        
        - `from`: used for getting the string representation of any value in RoLang. `io.println` and `io.print` internally use this to convert all values to string before printing them. Accepts a single value of any type.
        
        - `len`: Returns the number of characters in the string. Takes a single argument, the string.
        
        - `trim`: Trims the string based on the substr provided. Takes the original string as first argument and substr to trim as the second and trims the substr from original from left and right and returns a new string.
        
        - `trimSpace`: Like `trim` except for whitespaces only.
        
        - `split`: Splits the string based on the provided delimeter. Takes two argument, the string value and then the delimeter string
        
        - `splitSpace`: Like `split` but splits on whitespaces

        Example
        ```
        |> let name = "     Mona    Lisa    ";

        |> io.println(name);
             Mona    Lisa    
        |> name = strings.trimSpace(name);
        
        |> io.println(name);
        Mona    Lisa

        |> name = strings.splitSpace(name);

        |> io.println(name);
        ["Mona", "Lisa"]
        ```

    - `arrays`: This module deals with array manipulation
        - `len`: Takes an array as an argument and returns its length

        - `insert`: Takes three arguments, first the array, next the index at which to insert the element and third the element to insert. If index is out of bound throws an error

        - `erase`: Takes two arguments, first the array, next the index whose element is to be erased. Returns the erased element.

        - `push`: Takes two arguments, first the array, next the element and pushes this element to the back of the array.

        - `pop`: Takes a single argument, the array and pops the last element out from it and returns it.

        - `concat`: Takes variable number of arguments, each an array type and combines all their elements into a new array and returns the new array. Same as using `+` operator.

        - `copy`: Returns a new copy of the array instead of a reference.

        Example
        ```
        |> io.println(arrays.len([1, 2, 3]));
        3

        |> io.println(arrays.concat([1, 2], ["a", "b"]));
        [1, 2, "a", "b"]

        |> let a = [1, 2, 3];

        |> let b = a;

        |> arrays.push(a, 5);

        |> io.println(a);
        [1, 2, 3, 5]

        |> io.println(b);
        [1, 2, 3, 5]

        |> b = arrays.copy(a);

        |> arrays.pop(a);

        |> io.println(a);
        [1, 2, 3]

        |> io.println(b);
        [1, 2, 3, 5]
        ```

    - `maps`: This module deals with map types.
        
        - `len`: Takes a map as argument and returns the number of key, value pairs in it

        - `insert`: Takes three arguments, the map, the key and the value. If same key already exists it doesn't insert a new value instead returns false, else it creates a new key pair and returns true. To replace the old value, you can use the index `[]` operator

        - `erase`: Takes two arguments, the map and the key and attmepts to erase the key, value from the map if it exists and returns the value otherwise returns null

        - `concat`: Combines multiple maps into a single map. If more than one map contains the same key then the value of the last map becomes the value of the final map for that key, otherwise performs union operation of all maps. Same as using `+` operator.

        - `copy`: Like `arrays.copy` except for map types

        Example
        ```
        |> let m1 = {"a": 1, "b": 2};

        |> let m2 = {"a": 2, "c": 3};

        |> let m3 = m1 + m2;

        |> io.println(m3 == maps.concat(m1 + m2));
        true
        ```
    - `builtin`: This is not a module per-se, since all builtin functions are made available in global scope, so you do not need to use `builtin.<functionName>` but simply doing `<functionName>` is enough. Both works

        - `type`: Takes any element and returns a string denoting the type of value.

        Example
        ```
        |> io.println(type("Hello World"));
        string

        |> io.println(builtin.type([1, 2, 3]));
        array
        ```
    
## How to install

To build from source you need to have a Go compiler `verion >= 1.23`, install it from Go's official website and then clone this repo and build it
```bash
$ git clone https://github.com/RougherO/RoLang.git

$ cd RoLang

$ go build              # builds the interpreter

$ ./RoLang              # drops into repl mode
RoLang v0.3 Tree-Walk Interpreter
|> return;

$ ./RoLang example.ro   # interpretes the example.ro file
```

## Contributing

Contribution can be in any way or form, simply installing the language and testing it in ways to break it and creating new issues is also a significant contribution, additional ways you could help is writing more tests for different modules which are currently left as todos. If you happen to resolve any open issues or simply want to add a new feature, just create a PR on a new branch with your fix/feature as your branch name, also before making a PR make sure all current tests pass.

## Future

Since this is a hobby project, I don't get any incentive for developing this project other than my own personal enjoyment, however lately working on this for one month (and a previous attempt using LLVM and C++ for 3 months) had burnt me out a lot, so I'll be taking a break from this.

Some features that are missing and should be at the top of the list to get done when I come back or if someone wants to contribute, can check out the issues sections, other than that features that I myself would really like to see in the possible future releases are
- classes/structs
- proper module support for external `.ro` files
- a statically typed alternate for this language, which transpiles to C++ ( my favorite )