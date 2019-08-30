# Impact website

This is based on a lovely [material design template](https://github.com/joashp/material-design-template) from [Joash Pereira](http://joashpereira.com). It is a single page website with all content contained in `index.html`.

The CSS framework is [Materialize](http://materializecss.com/). We also uses icons from [Font Awesome](http://fontawesome.io/icons/). See those links for docs or use Google to find related tutorials.

# Setting up on windows

This project uses [Jekyll](https://jekyllrb.com) which is closely integrated into GitHub Pages. You will need [Git Bash](https://git-scm.com/downloads) or [some other](https://medium.com/@borekb/zsh-via-msys2-on-windows-3964a943b1ce) suitable MSYS2 setup as well as [ruby](https://www.ruby-lang.org/en/downloads) version 2.1.0 or higher installed and on your `PATH`.

- Open Git Bash (or some other MSYS2 environment with git and ruby installed)
- Verify git is installed by running `git --version`
- Verify ruby is installed by running `ruby --version`
- Install [bundler](https://bundler.io) if you haven't already by running `gem install bundler`
- If you haven't already, clone this repo (or your fork of it)
- `cd` into the directory you cloned the repo to
- run `bundle install` to install the dev dependancies this project requires

# Setting up on Linux

Even easier! Just install ruby via your package manager, ensure bundler is installed (`gem install bundler`) and run `bundle install` in your local git repo.

# Running the site locally

Once you've set up your local dev environment by running `bundle install` (see above) you can "run" the site on your machine by running `bundle exec jekyll serve`. This command will output something like

```
Configuration file: ~/Development/website/_config.yml
            Source: ~/Development/website
       Destination: ~/Development/website/_site
 Incremental build: disabled. Enable with --incremental
      Generating...
                    done in 2.809 seconds.
 Auto-regeneration: enabled for '~/Development/website'
LiveReload address: http://127.0.0.1:35729
    Server address: http://127.0.0.1:4000
  Server running... press ctrl-c to stop.
```

Navigating to the "server address" (in the above example, `http://127.0.0.1:4000`) in a web browser will show you the website complete with any local changes you have made. Jekyll will watch for changes and will rebuild and reload if you modify any of the source files.

# Keeping your environment up to date

- From the repo directory run `bundle update`

# Livereload errors on Windows
If you encounter an error such as
```
Unable to load the EventMachine C extension; To use the pure-ruby reactor, require 'em/pure_ruby'
Traceback (most recent call last):
        22: from C:/Ruby25-x64/bin/jekyll:23:in `<main>'
        21: from C:/Ruby25-x64/bin/jekyll:23:in `load'
[...]
C:/Ruby25-x64/lib/ruby/gems/2.5.0/gems/eventmachine-1.2.7-x64-mingw32/lib/rubyeventmachine.rb:2:in `require': cannot load such file -- 2.5/rubyeventmachine (LoadError)
```
Try running
```
gem uninstall eventmachine
gem install eventmachine --platform ruby
```
See [this comment](https://github.com/RobertDeRose/jekyll-livereload/issues/18#issuecomment-353386266) for more details.