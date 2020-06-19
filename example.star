def run():
    for n in range(100):
        writeline("what is your name?")
        name = readline()
        writeline()
        writeline("hello " + name)
        writeline("enter a number:")
        [number] = expect(r"\d")
        writeline()
        writeline("hello " + number + ". again? [y/n]")

        [more] = expect(r"[yn]")
        if more == "n":
            break

run()