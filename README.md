mistify-operator-admin
======================

The "Operator Admin API" is used by the system administrators to control the Mistify system. This includes such things as adding or removing users, adding or removing hypervisors, and general configuraion. Higher level functionality, like creating guests, will not handled here.

The API is implemented as a simple HTTP service in Go. The datastore is a redundant, replicated "cluster" of Postgres servers.

All data is datacenter specific unless otherwise specified. All ids are uuids unless otherwise specified. All entities except for config have a metadata `map[string]string`

List results are JSON arrays of of the particular objects (e.g. GET /hypervisors returns an array of Hypervisor objects). Single get and create results areturn the particular object. Relation results return empty objects.

## Testing

Run `make test`. This will create the necessary test database and db user, run the tests for the various subpackages, and then clean up the database and db user. The clean happens at the beginning and the end, so there is no need to explicitly run `make test_clean` after a failed test run before the next.

## API Endpoints

### Config
Configuration is stored with namespaced keys. Keys will use set values and fall back to defaults. Custom namespaces and keys can be created and deleted, while core namespaces and keys can only be unset (the defaults remain).

* `/config`
    * `GET` - Get the full config set
    * `PUT` - Set the full config set
    * `PATCH` - Set part of the config
* `/config/{namespace}`
    * `GET` - Get a namespace config set
    * `PUT` - Set a namespace config set
    * `DELETE` - Delete a namespace config set
* `/config/{namespace}/{key}`
    * `DELETE` - Delete a key

### Flavors
Flavors represent a desired set of system resources, similar to AWS EC2's instance types `m3.medium` or `c3.2xlarge`

* `/flavors`
    * `GET` - Get a list of flavors
    * `POST` - Create a new flavor
* `/flavors/{flavorID}`
    * `GET` - Get a flavor
    * `PATCH` - Update a flavor
    * `DELETE` - Remove a flavor

### Hypervisors
Hypervisors run on physical machines and manage the virtual guests.

* `/hypervisors`
    * `GET` - Get a list of hypervisors
    * `POST` - Register a hypervisor
* `/hypervisors/{hypervisorID}`
    * `GET` - Get a hypervisor
    * `PATCH` - Update a hypervisor
    * `DELETE` - Deregister a hypervisor
* `/hypervisors/{hypervisorID}/ipranges`
    * `GET` - Get a list of ipranges related to a hypervisor
    * `PUT` - Set a list of ipranges related to a hypervisor
* `/hypervisors/{hypervisorID}/ipranges/{iprangeID}`
    * `PUT` - Associate a hypervisor with an iprange
    * `DELETE` - Disassociate a hypervisor from an iprange

### IP Ranges
IP ranges are configured ip blocks that are associated with hypervisors for guests to get allocated from.

* `/ipranges`
    * `GET` - Get a list of IP ranges
    * `POST` - Create an IP range
* `/ipranges/{iprangeID}`
    * `GET` - Get an IP range
    * `PATCH` - Update an IP range
    * `DELETE` - Delete an IP range
* `/ipranges/{iprangeID}/hypervisors`
    * `GET` - Get a list of hypervisors associated with an IP range
    * `PUT` - Set a list of hypervisors associated with an IP range
* `/ipranges/{iprangeID}/hypervisors/{hypervisorID}`
    * `PUT` - Associate an IP range with a hypervisor
    * `DELETE` - Disassociate an IP range from a hypervisor
* `/ipranges/{iprangeID}/network`
    * `GET` - Get the network associated with an IP range
* `/ipranges/{iprangeID}/network/{networkID}`
    * `PUT` - Set the network associated with an IP range
    * `DELETE` - Disassociate the network associated with an IP range

### Networks
Networks are named sets of IP Ranges.

* `/networks`
    * `GET` - Get a list of networks
    * `POST` - Create a network
* `/networks/{networkID}`
    * `GET` - Get a network
    * `PATCH` - Update a network
    * `DELETE` - Delete a network
* `/networks/{networkID}/ipranges`
    * `GET` - Get a list of IP ranges associated with a network
    * `PUT` - Set a list of IP ranges associated with a network
* `/networks/{networkID}/ipranges/{iprangeID}`
    * `PUT` - Associate a network with an IP range
    * `DELETE` - Disassociate a network from an IP range

### Permissions
Permissions are allowed actions on entities for services. Permissions are associated with projects, with users in those projects being granted the associated permissions.

* `/permissions`
    * `GET` - Get a list of permissions
    * `POST` - Create a permission
* `/permissions/{permissionID}`
    * `GET` - Get a permission
    * `PATCH` - Update a permission
    * `DELETE` - Delete a permission
* `/permissions/{permissionID}/projects`
    * `GET` - Get a list of projects associated with a permission
    * `PUT` - Set a list of projects associated with a permission
* `/permissions/{permissionID}/projects/{projectID}`
    * `GET` - Associate a project with a permission
    * `DELETE` - Disassociate a project from a permission

### Projects
Projects are groups of users and are what take ownership of entities like guests.

* `/projects`
    * `GET` - Get a list of projects
    * `POST` - Create a project
* `/projects/{projectID}`
    * `GET` - Get a project
    * `PATCH` - Update a project
    * `DELETE` - Delete a project
* `/projects/{projectID}/users`
    * `GET` - Get a list of users associated with a project
    * `PUT` - Set a list of users associated with a project
* `/projects/{projectID/users/{userID}`
    * `PUT` - Associate a project with a user
    * `DELETE` - Disassociate a project from a user
* `/projects/{projectID}/permissions`
    * `GET` - Get a list of permissions associated with a project
    * `PUT` - Set a list of permissions associated with a project
* `/projects/{projectID}/permissions/{permissionID}`
    * `PUT` - Associate a project with a permission
    * `DELETE` - Disassociate a project from a permission

### Users
Users in the system.

* `/users`
    * `GET` - Get a list of users
    * `POST` - Create a user
* `/users/{userID}`
    * `GET` - Get a user
    * `PATCH` - Update a user
    * `DELETE` - Remove a user
* `/users/{userID}/projects`
    * `GET` - Get a list of projects the user is related to
    * `PUT` - Set a list of projects the user is related to
* `/users/{userID}/projects/{projectID}`
    * `PUT` - Associate a user with a project
    * `DELETE` - Disassociate a user from a project

## Contributing

See the [contributing guidelines](./CONTRIBUTING.md)
