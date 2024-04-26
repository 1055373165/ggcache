package service

/*
Picker is responsible for finding which node the query request for key should be sent to.（using consistent hashing algorithm）
*/
type Picker interface {
	Pick(key string) (Fetcher, bool)
}

/*
Fetcher is responsible for querying the value of key in the specified group cache.

every distributed kv node should be implement this interface.
*/
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

/*
Retriever interface for retrieving data from backend database.

When the value of the key cannot be queried from the node's group cache,

the system needs to provide alternative options, i.e., go to the back-end database to query the key's value.
*/
type Retriever interface {
	retrieve(string) ([]byte, error)
}

/*
By adding a method retrieve to RetrieveFunc, an instance of this function type can be used as an implementation of the Retriever interface;

this is a classic adapter pattern, adapting a function to an interface.

Instead of a complete structure to implement the interface, only a function is needed to meet the requirements of the interface, which is a common simplification technique in go,

making the code more concise and easy to understand; this pattern allows for quick customization to change or inject the data retrieval strategy, especially for scenarios that require a high degree of flexibility and dynamic data processing.

For example, if you are building a microservice architecture that needs to retrieve data from multiple data sources, you can easily switch between different data retrieval strategies without affecting other business logic.
*/
type RetrieveFunc func(key string) ([]byte, error)

/*
RetrieveFunc implements the retrieve method, that is, implements the Retriver interface so that any anonymous function func through RetrieverFunc ( func ) forced type conversion, the ability to achieve the Retriver interface.
This is also reflected in the gin framework inside the HandlerFunc type encapsulation anonymous function, http type handler forced conversion can be directly used as gin Handler.
*/
func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}
