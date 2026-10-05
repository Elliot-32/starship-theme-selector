Output docs/demo.gif

Require ssts
Require starship

Set Shell "bash"
Set FontFamily "FiraCode Nerd Font"
Set FontSize 22
Set Width 1200
Set Height 720
Set Padding 18
Set Framerate 30
Set TypingSpeed 45ms

Type "ssts"
Enter
Wait+Screen@10s /Starship Theme Selector/
Sleep 1s

Type "/"
Type@90ms "pure-preset"
Wait+Screen@10s /Preview — pure-preset/
Sleep 1s
Enter
Sleep 500ms
Enter
Wait+Screen@10s /Applied built-in Starship theme: pure-preset/
Sleep 2s
