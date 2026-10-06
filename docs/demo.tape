Output docs/demo.gif

Require ssts
Require starship

Set Shell "bash"
Set FontFamily "JetBrains Mono"
Set FontSize 22
Set Width 1200
Set Height 720
Set Padding 18
Set Framerate 30
Set TypingSpeed 45ms

Env XDG_CONFIG_HOME "/tmp/ssts-vhs-demo-config"
Env STARSHIP_CONFIG "/tmp/ssts-vhs-demo-starship.toml"

Type "ssts"
Hide
Enter
Wait+Screen@10s /Starship Theme Selector/
Show
Sleep 1s

Type "/"
Type@90ms "pure-preset"
Hide
Wait+Screen@10s /Preview — pure-preset/
Show
Sleep 1s
Enter
Sleep 500ms
Hide
Enter
Wait+Screen@10s /Applied built-in Starship theme: pure-preset/
Show
Sleep 2s
