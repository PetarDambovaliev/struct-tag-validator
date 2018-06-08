# struct-tag-validator

This is a general struct validator

```
m := NewValidator("path/to/your/structs")
```

Add a specific tags to be validated or use * for all
Adding default processors (validators)

```
m.AddDefaultProcessors("db", "json")
```


Add your own processor

```
m.AddProcessor("db", func(tag *Tag) []error {
		errs := []error{}

		if len(tag.GetValue()) > 2 {
			errs = append(errs, errors.New("Too long"))
		}

		return errs
	})
```
  
  
  Run the validator
  
 ```
 errs := m.Run()
 ```
