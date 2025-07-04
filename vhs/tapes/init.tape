Output output/init.gif
Output output/init.txt

Set WindowBar Colorful

Set Theme OneHalfLight

Set Width 1500
Set Height 1000

Set PlaybackSpeed 0.75
Set Framerate 30

Env COLORTERM truecolor

Hide
Type "rm -f /root/.conjurrc"
Enter
Type "rm -f /root/conjur-server.pem"
Enter
Type "clear"
Enter
Show

# Start the init command
Type "conjur init"
Enter
Sleep 2s

# Choose environment 'enterprise'
Down
Sleep 1s
Enter
Sleep 500ms

# Provide the URL for the Conjur server
Type "https://example.com"
Enter
Sleep 2s

# Provide account name
Type "conjur"
Sleep 1s
Enter

Sleep 5s
