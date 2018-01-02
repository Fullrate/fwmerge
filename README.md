fwmerge
=======
_- a firewall ruleset renderer_

`fwmerge` is a firewall ruleset renderer for firewalls that support a
table/chain/rule structure(like `iptables` and `nftables`). It takes YAML files
as inputs and outputs a ruleset that can be loaded into the given firewall. Each
rule is tagged with a priority allowing `fwmerge` to merge chains and sort the
rules. The final ruleset is then output in the requested format. `fwmerge`
doesn't know about specific rules, and cannot translate between different
firewall syntaxes.

A sample rule file YAML could look like the following:

```yaml
filter:
  INPUT:
    - policy: DROP
    - 10 allow ICMP: -p icmp -j ACCEPT
    - 10 allow all on loopback: -i lo -j ACCEPT
    - 10 allow SSH: -p tcp --dport 22 -j ACCEPT
  testchain: unmanaged
```

`fwmerge` can set the policy for chains that support it using the policy tag.
Note that the policy tag has no prioirty, the last policy set will win. If no
policy is set, the default will be set to **ACCEPT**.

`fwmerge` can also create unmanaged chains. These are chains that fwmerge will
ask the firewall to create, but it won't output rules to populate the chain.
This allows other applications to manage these chains without interference.

The rules are specified as either:

  - `<priority>: <rule>`
  - `<priority> <comment>: <rule>`

The priority is used for sorting, the comment is ignored, and the rule is output
verbatim into the ruleset. The rule must be convertable to a string.

Generators
----------

For now the only supported generator is `iptables`. This generator will output a
ruleset that can be piped to `iptables-restore`.

For the `iptables` generator, all rules must be valid `iptables` match/target
specifications. Do not include the actual insertion command(e.g. `-A`).

Ruleset merging
---------------

Multiple rulesets can be specified as inputs to `fwmerge`, and all rulesets will
be combined before the final ruleset is output. Merging is done very simply by
merging all tables/chains together, and stably sorting the final ruleset
chains by priority. As an example, given the following two rulesets:

```yaml
# Ruleset 1
filter:
  INPUT:
    - 10 allow SSH: -p tcp --dport 22 -j ACCEPT 
    - 10 allow DNS: -p udp --dport 53 -j ACCEPT 

# Ruleset 2
filter:
  INPUT:
    - 5 allow ICMP: -p icmp -j ACCEPT 
```

`fwmerge` will combine this into the following ruleset:

```yaml
# Ruleset 1
filter:
  INPUT:
    - 5 allow ICMP: -p icmp -j ACCEPT 
    - 10 allow SSH: -p tcp --dport 22 -j ACCEPT 
    - 10 allow DNS: -p udp --dport 53 -j ACCEPT 
```

Note that the sort is stable, which means that rules within a single file will
always be put in the same order in the output, if they have the same priority.
This works across files as well, but users are advised not to rely on any
behavior related to file ordering, and solely use the priority system.

Unmanaged chains
----------------

`fwmerge` supports the notion of an *unmanaged chain*. These are chains that
`fwmerge` knows the existance of, and generates commands for creation of, but
the contents of the chain is not tracked and is never modified in the ruleset
output by `fwmerge`. This is useful for tools like
[fail2ban](https://www.fail2ban.org/) that work by updating `iptables` chains
dynamically.

An unmanaged chain is specified in the ruleset by the word `unmanaged` used as
the value of chain, e.g.:

```yaml
filter:
  fail2ban-ssh: unmanaged
```

This will ensure that the fail2ban-ssh chains exists, but otherwise will not
touch the contents of the chain.

Chain policy
------------

Chain policies can be specified with a rule like:
```yaml
filter:
  INPUT:
    - policy: DROP
```

The last set policy will win, so users should take care to merge files in the
right order if this property is to be used. It is advised that only a single
file sets the policy. Alternatively, if files are to be collected via a shell
glob, to name files in the format `NN_name.yaml` where `NN` denotes a number.

Motivation
----------

`fwmerge` was designed to be used together with configuration management systems
like [Salt](https://saltstack.com/) or [Puppet](https://puppet.com/).

With `fwmerge`, these systems can manage a single directory with firewall
rulesets instead of statefully trying to manage the ruleset. The final ruleset
can then be assembled by `fwmerge`. It is recommended to wrap `fwmerge` in a
service for easy firewall reloads.
