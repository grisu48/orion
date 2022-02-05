# orion

**Alpha state package!** *It has not been hardened and contains probably bugs, but it is working.*

orion is a secure yet minimalistic gemini server written in go

## Usage

Compile the `orion` binary:

    make

Then edit the configuration file `orion.conf` to your wishes and launch the program

    ./orion -config orion.conf

### Create self-signed certificates

To create self-signed certificates for quick testing, you can use the following recipe

    make cert

## Credits

* This project was inspired by the [titan2](https://gitlab.com/lostleonardo/titan2) minimalistic Gemini server written by lostleonardo.
