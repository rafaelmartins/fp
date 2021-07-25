# fp

A simple filebin media player.


## Config

Path: `~/.fp.yml` or `$FP_CONFIG`

```yaml
handlers:
    video/*:
        - mpv
        - --fs
    image/*:
        - feh
        - --auto-zoom
        - --fullscreen
        - --hide-pointer

aliases:
    foo: https://example.org/foo
    bar: https://example.org/bar
```


## Bash Completion

```bash
complete -C /path/to/fp fp
```
