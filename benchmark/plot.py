#!/usr/bin/env python3

import matplotlib.pyplot as plt
import pandas as pd
import seaborn as sns
import glob


def main():
    path = "./measurements"
    files = glob.glob(path + "/*.csv")

    # combining measurements from different clients into one data frame
    list = []
    for f in files:
        df = pd.read_csv(f, names=["time-to-certificate"])
        list.append(df)
    df = pd.concat(list)


    bplot = sns.boxplot(y="time-to-certificate", 
                    data=df,
                    width=0.2,
                    palette="colorblind")

    # show plot
    plt.show()


if __name__ == '__main__':
    main()
