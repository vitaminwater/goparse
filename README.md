Introduction
===

This is juste a simple and incomplete abstraction over [Parse](http://parse.com) Rest API.

API
===

The goparse API is based on the `Object` type.
An `Object` is just a key-value store, it exposes `Get` and `Set`
methods with different flavours of types.

`Object` instances is Marshable.

Read source comments for a list of exposed methods.

---

The `ClassObject` is what you use to interact with documents in a Parse
Collection.

the `ClassObject` inherits from `Object`, and adds a `Save` method to
send to Parse, and `Delete` to delete the entry.
