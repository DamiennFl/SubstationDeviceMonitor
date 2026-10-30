## Go Notes

Default time in Go: 2009-11-10 23:00:00 UTC, comes from the day Go was created

- Go programs made up of packages
- Programs start running in package main

### Packages
Package name is the same as the last element of the import path. "math/rand" package comprises files that begin with statement package rand.
So this is a way to group similar files into one package.

The Go sandbox is deterministic? Is Go deterministic? No it's not. Just Go's math.rand is deterministic, based on a seed.

### Types

Type declarations come after the variable name in parameters.

Multiple variables with same type can be declared in the same line, e.g. `x, y int`.

### Functions and variables
Function can return any number of results.

Return values may be named. Treated as variables at the top of the function. Names are usually used to document the meaning of the return values.

A return statement without arguments returns the named return values. This is a naked return. Use only in short functions, can harm readability. If values for return vars are not set then the default is returned.

var declares list of variables, type at end, default vals. Can be at package or function level. Can include initializers, one per variable. type can be omitted, will take type of the initializer.

:= short assignment statement for implicit types, e.g. basic types: strings, bools, ints, floats, etc. ONLY IN FUNCTIONS, since outside every statement has to start with keyword.

const for constants. Constants can also be type inferred.

### Loops and Defer

for loop is basically default, except no parentheses, brackets required. variable init and post statements are optional, not comparison obviously.

while loop is done by using for with only comparison. infinite loop:

for {
}

if statements also no parentheses, no colon, brackets are required though. Vars in if statement declaration are present in if and else statement.

switch is the same. No break needed since Go will stop after the first (in order) successful case. `switch true` can be implemented with switch with no condition, e.g. for long if-then-else chains.

`defer` statement defers the execution of a function until the surrounding function returns. Deferred call's arguments are evaluated immediately but not executed until surrounding function returns. They are pushed onto a stack, when a function returns, its deferred calls are executed in last in first out order.

### Pointers

Go has pointers. type `*T` is a pointer to a `T` value. Its zero value is nil.

`var p *int`

& operator generates pointer to operand.
```
i := 42
p = &i      p is now pointer to i (42)

fmt.Println(*p) // read i through the pointer p
*p = 21         // set i through the pointer p
```

Structs are collections of fields.
Struct fields are accessed using dot (like attributes in Python)

No need for explicit dereference when using a pointer to get fields
```
type Vertex struct {
	X int
	Y int
}

func main() {
	v := Vertex{1, 2}
	p := &v
	p.X = 1e9
	fmt.Println(v)
}
```

Can also do struct literals, and pointers to struct literals defined directly. E.g. `v1 = Vertex{1, 2}`

### Arrays

Arrays are `[n]T` where n is size, T is type.

Array length is part of its type, so arrays cannot be resized.

### Slices

Slices are dynamically sized, flexible view into elements of an array. Slices are much more common than arrays.

`[]T` is a slice of elements of type `T`.

`a[low:high]`, half open range so like Python. Last is exclusive, first is inclusive.

Slices do not store data, just describe section of underlying array. Changing slice elements changes underlying array.

A slice literal is hte same as an array literal, without a length.
```
[3]bool{true, true, false} and []bool{true, true, false} will both make an array, however the second will make a slice which references it.
```

You can slice only one side as well, e.g. `[:1] or [1:]`, like Python.

Slices have length AND capacity. Length of a slice is the number of elements it contains. Capacity is the number of elements in the underlying array, counting from the first element in the slice. Can both be obtained using `len(s)` and `cap(s)` where s is the slice.

Zero value of slice is nil. Nil slice has length and capacity of 0, and no underlying array.

Slices can be created using `make` function, which allows for creating dynamically sized arrays. `make` allocates a zeroed array and returns a slice that refers to that array.
```
a := make([]int, 5) // len(a) = 5

b := make([]int, 5, 5) // len(a) = 5, cap(a) = 5
```

Slices can have any type, including other slices.

.append works same as Python, same in undertlying logic as well. Append will add one or more values to end of slice, and will re-allocate larger array if necessary.

Range form of for loop iterates over slice or map. When ranging over slice, two values are returned at each iteration: index, and copy of element at that index. So like python enumeration

Can throw away index or value by using `_`, same as Python

```
package main

import "golang.org/x/tour/pic"

func Pic(dx, dy int) [][]uint8 {
	ret := make([][]uint8,0, dy)
	
	for i := 0; i < dy; i++ {
		curr := make([]uint8, 0, dx)
		for j := 0; j < dx; j++ {
			val := uint8(i^j)
			curr = append(curr, val)
		}
		ret = append(ret, curr)
	}
	return ret
}

func main() {
	pic.Show(Pic)
}
```

### Maps

Maps map keys to values. Zero value of a map is nil. A nil map has no keys nor can keys be added. The `make` function returns a map of the given type, initialized and ready for use.

```
package main

import "fmt"

type Vertex struct {
	Lat, Long float64
}

var m map[string]Vertex

func main() {
	m = make(map[string]Vertex)
	m["Bell Labs"] = Vertex{
		40.68433, -74.39967,
	}
	fmt.Println(m["Bell Labs"])
}
```

Map literals exist, but need keys directly defined. If the top-level type is just a type name, no need to redefine it in the key value pairs definition, only necessary in the main map literal definition.

Mutating maps:
```
m[key] = elem // Insert or update element

elem = m[key] // Retrieve an element

delete(m, key) // Delete an element

elem, ok = m[key] // Test key is present with two-value assignment. if ok is true, key is present, and elem will be the element, or zero value for elem's type.
```

WordCount example
```
package main

import (
	"golang.org/x/tour/wc"
	"strings"
)

func WordCount(s string) map[string]int {
	words := strings.Split(s, " ")
	m := make(map[string]int)
	
	for _, word := range(words) {
		_, ok := m[word]
		if ok {
			m[word] += 1
		} else {
			m[word] = 1
		}
	}
	return m
}

func main() {
	wc.Test(WordCount)
}
```

### Function Values

Function values can be used as values as well.

`func compute(fn func(float64, float64) float64) float64 {` compute is a function, which takes in `fn` as a parameter, which is itself a function which takes two float64s.

A function closure is a function value that references variables from outside its body. The function may access and assign to the reference variables, in this sense the function is bound to the variables.

```
func adder() func(int) int { // Adder itself has no parameters.
	sum := 0
	return func(x int) int { // The function closure takes in a value, modifies the number, and returns it.
		sum += x
		return sum
	}
}

pos = adder() // This is an int, since adder returns an int.

pos(1) will add 1 to 0, and pos will store that value. If pos(1) is called again, it will increment to 2.
```

### Methods and Receivers

Go does not have classes, however you can define methods on types
A method is a function with a special receiver argument
Receiver appears in its own argument list between func keyword and the method name.
```
type Vertex struct {
	X, Y float64
}

func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func main() {
	v := Vertex{3, 4}
	fmt.Println(v.Abs()) // You can directly use .Abs on v because the Abs function has a receiver v of type Vertex.
}
```

You can only declare a method with a receiver whose type is defined in the same package as the method. You cannot declare a method with a receiver whose type is defined in another package, such as built-in types.

You can declare methods with pointer receivers. This allows you to modify the value to which the receiver pointer points to. These are more common than value receivers.

A function with a pointer argument must take a pointer, but a function with a pointer receiver can take pointer or value. However, the underlying functionality still uses a pointer; Go interprets it with a pointer automatically, since the function has a pointer receiver. Works in reverse order as well.

Two reasons to use a pointer receiver over a value receiver. First, so a method can modify what the pointer references. Second, to avoid copying the value on each method call, and instead reference the same memory each call.

### Interfaces

Interface type is defined as a set of method signatures. A value of interface type can hold any value that implements those methods.

Interfaces are implemented implicitly. A type implements an interface by implementing its methods. No explicit declaration of intent, no "implements" keyword. Implicit interfaces decouple the definition of an interface from its implementation. It can appear in any package without prearrangement. 

```
type I interface {
	M()
}

type T struct {
	S string
}

// This method means type T implements the interface I,
// but we don't need to explicitly declare that it does so.
func (t T) M() {
	fmt.Println(t.S)
}

func main() {
	var i I = T{"hello"}
	i.M()
}
```

Under the hood, interface values can be thought of as a tuple of a value and a concrete type. `(value, type)`

An interface value holds a value of a specific underlying concrete type. Calling a method on an interface value executes the method of the same name on its underlying type.

If the concrete value inside the interface itself is nil, the method will be called with a nill receiver. An interface that holds a nil concrete value is itself non-nil.

A nil interface holds neither a value nor a concrete type. Calling a method on a nil interface is a runtime error because there is no type inside the interface tuple to indicate which concrete method to call.

### Type Assertions

Type assertion provides access to an interface value's underlying concrete value.
```
t := i.(T) // This statement asserts that the interface value i holds the concrete type T and assigns the underlying T value to the variable t. 
```

If `i` does not hold a `T`, this will cause an error (panic). This is testable: `t, ok := i.(T)`. If `i` holds a T, then `t` will will be the underlying value and ok will be true. Same syntax as reading from a map.

### Type Switches

A type switch is a construct that allows multiple type assertions in a series. Like a regular switch statement, but cases specify type. Compared against type of value held by given interface value. Declaration of type switch has the same syntax as type assertion `i.(T)` but T is replaced with keyword `type`.

### Stringers

A Stringer is a type that can describe itself as a string. Packages look for this interface to print values.

### Errors

Programs express error state with error values. Error type is a built-in interface.

```
type error interface {
	Error() string
}
```

Functions often return an `error` value, calling code should handle errors by testing whether error equals nil. A nil error denotes success; a non-nil error denotes failure.

```
type ErrNegativeSqrt float64

func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprintf("cannot Sqrt negative number: %v", float64(e))
}

func Sqrt(x float64) (float64, error) {
	if x < 0 {
		return x, ErrNegativeSqrt(x)
	}
	z := float64(1)
	prev := z
	for {
		prev = z
		z -= (z*z - x) / (2 * z)
		if z-prev < 0.0000000000001 {
			return z, nil
		}
	}
}
```

### Readers

`io` package specifies `io.Reader` interface, represents the rear end of a stream of data.

`io.Reader` has a Read method: func (T) Read(b []byte) (n int, err error)

Read populates the given byte slice with data and returns the number of bytes populated and an error value. Returns an `io.EOF` error when the stream ends.

TODO: Reader exercises

### Type parameters

`func Index[T comparable](s []T , x T) int`: s is a slice of any type T that fulfills the built-in constraint `Comparable` `x` is a value of the same type.

Also supports generic types. A type can be parameterized with a type parameter, which could be useful for implementing generic data structures. 

```
// List represents a singly-linked list that holds
// values of any type.
type List[T any] struct {
	next *List[T]
	val  T
}
```

## Concurrency

Go has built in concurrency features as part of the core language.

### Goroutines

Lightweight thread managed by Go runtime. `go f(x, y, z)`. Start a new goroutine by running f(x, y, z). The eval of `f`, `x`, `y`, and `z` happens in this goroutine, and f is executed in the new goroutine.

Goroutines run in the same address space, so access to shared memory must be synchronized. The `sync` package provides useful primitives.

### Channels

Channels are a typed conduit through which you can send and receive values with the channel operator `<-`. Data flows in the direction of the arrow. Channels must be made before use. `ch := make(chan int)`. By default, sends and receives block until the other side is ready. Allows goroutines to synchronize without explicit locks or condition variables. 
```
func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
```

Channels can be buffered. Provide the buffer length as the second argument to `make` to initialize a buffered channel. Data is sent to the channel buffer, and blocks once the buffer is full. Receives block when the buffer is empty (no data to receive).

A sender can close a channel to indicate no more values will be sent. `v, ok := <-ch` can be used to confirm the channel is closed. Only the sender should close a channel, never the receiver. Sending on a closed channel will cause a panic. Closing is not strictly necessary, rather should be used to tell receiver no more values are coming. 

### Select

`select` statement lets a goroutine wait on multiple communication operations. A `select` blocks until one of its cases can run, then executes that case. It chooses one at random if multiple are ready. 

The default case runs if no other case is ready.

### Mutex

Go provides mutex with standard library, specifically with `Lock` and `Unlock` methods. 
```
// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.v[key]++
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.v[key]
}
```

### Struct tags

Used to define how external libraries manipulate, serialize, or validate data.

```
type User struct {
    ID    int    `json:"id" db:"user_id"`
    Email string `json:"email,omitempty" validate:"required,email"`
}
```