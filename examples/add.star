def run():
    total = 0
    for _ in forever():
        writeline("enter a number:")
        [number] = matchline(r"\d+")

        total += int(number)

        writeline("current total: " + str(total) + "")
        writeline("again? [y/n]")

        [more] = matchline(r"[yn]")
        if more == "n":
            break

run()