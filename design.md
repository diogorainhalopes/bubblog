
Entry Model
    state
    menu    tea.Model   (List)
    log     tea.Model   (Multiple views)


UI design
    Initial screen:
        Recent log files list
        Select log file
        Select log file from directory

    Main screen:
        Full log file on the left side
        One line text input on the bottom
        Filtered log file on the right side 



---

Cache
    ~/.cache/golog/cache.json
```
{
    "files": [
        {
            "path": "/home/user/logfile.log",
        }
    ]
}
```


Possible future additions:
    Tabs for multiple log files
    Caching