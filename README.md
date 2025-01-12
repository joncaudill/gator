# gator
## an rss aggregator project

a simple rss aggregator written in go

you will need to have postgresql and go installed in order to use this aggregator

you will need to set up a config file in your home directory for this to work.

the config file needs to be named .gatorconfig.json
the contents should be similar to the following

`{"db_url":"postgres://username:password@localhost:5432/gator?sslmode=disable","current_user_name":""}`

you will need to replace "username" and "password" with your postgresql username and password.

to install the software, navigate to the root of where you installed the software and type:

`go install`

at that point, gator should be installed.

to use:

gator *command* *parameters*

the commands available are:

- login *username*  - makes *username* the currently active profile
- register *username* - creates a profile for *username* and makes them the currently active profile
- reset - deletes all all data from gator and "factory resets" it.  **this cannot be undone**
- users - lists all profiles that have been created for the app
- agg *time* - goes out and re-aggregates all rss feeds that has been added to the app.  *time* should be a number followed by a unit in "h" for hours and "m" for minutes (e.g. "1h"). It will re-fetch all of the subscribed feeds every *time* interval.  **do not** use a very low time value here as it will likely upset the site owner and they may ban you from the site.  By default, the minimum time value allowed is 10m.  If you try to use a value lower than this, it will make the time value 10m.   Depending on the site, this may still be too low a value.  This is best run in another terminal, as it will keep running until stopped with **ctrl-c**. 
- addfeed *name* *url* - adds a feed to the app and subscribes the current profile to it. *name* is the name of the site in quotes, and *url* is the url for the site in quotes.
- feeds shows a list of all feeds that have been added to the app
-follow *url* adds the feed with the url *url* to the current profile's list of feeds that they follow
-following shows a list of all feeds the current profile is following
-unfollow *url* unfollows a feed with the url *url* from the list of feeds the current profile is following
-posts *num* shows the most recent *num* of posts from the feeds the current profile is following.   If *num* is not provided, it defaults to 2.





