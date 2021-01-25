# xproxy

## add to feeds.conf.default

```
src-git xproxy https://github.com/ccxp/xdns
```

## run

```
./scripts/feeds update xproxy
./scripts/feeds install xproxy
make menuconfig
```

## compile

```
make package/feeds/xproxy/xdns/compile V=99
```

