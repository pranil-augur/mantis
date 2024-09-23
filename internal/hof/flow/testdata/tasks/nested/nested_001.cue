package test

import (
	"test.ai/vars"
	nested1 "test.ai/nested"
)

@flow()
nested: {
	tasks: {
		get: {FOO: _ } @task(os.Getenv)
		out: { text: get.FOO + "\(vars.region)" + " here \n"} @task(os.Stdout)  
		foo: get.FOO
	}

	out: {text: tasks.get.FOO + "Hi there \n"} @task(os.Stdout)
	foo: nested1.install_mongodb
}

 

