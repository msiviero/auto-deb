# Auto deb
Experimental helper for building deb packages for long running executables managed by systemd on debian / ubuntu

## Usage
```
auto-deb -c ./debian.yml -o ./build-linux -v 0.1
```

The -v flag is optional and used to override the config version and meant specifically for CI systems