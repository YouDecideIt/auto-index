# AutoIndex

AutoIndex is a project for [TiDB Hackathon 2022](https://tidb.net/events/hackathon2022). 

We are aimed to implement a robust automatic index recommendation and TiFlash copy recommendation system. 

See [Our Chinese RFC here](https://gist.github.com/LittleFall/7f0ddfb2dd6d029e06d22760f6eb1de1).

# Run

compile

```sh
make default
```


run

```sh
./bin/auto-index --config.file=autoindex.yaml
```


# Contributors

Team YouDecideIt(”你说了算“ in Chinese)

[@littlefall](https://github.com/littlefall), for most code.

[@SunRunAway](https://github.com/SunRunAway), for the project idea, project design, and the `what-if` interface in TiDB.

[@liubog2008](https://github.com/liubog2008), for the hardest cloud deployment and traffic copy parts.

# Thanks

Thanks to https://github.com/pingcap/tidb and it's community, let's happy hacking!
