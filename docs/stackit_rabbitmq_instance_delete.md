## stackit rabbitmq instance delete

Deletes an RabbitMQ instance

### Synopsis

Deletes an RabbitMQ instance.

```
stackit rabbitmq instance delete INSTANCE_ID [flags]
```

### Examples

```
  Delete an RabbitMQ instance with ID "xxx"
  $ stackit rabbitmq instance delete xxx
```

### Options

```
  -h, --help   Help for "stackit rabbitmq instance delete"
```

### Options inherited from parent commands

```
  -y, --assume-yes             If set, skips all confirmation prompts
      --async                  If set, runs the command asynchronously
  -o, --output-format string   Output format, one of ["json" "pretty"]
  -p, --project-id string      Project ID
```

### SEE ALSO

* [stackit rabbitmq instance](./stackit_rabbitmq_instance.md)	 - Provides functionality for RabbitMQ instances

