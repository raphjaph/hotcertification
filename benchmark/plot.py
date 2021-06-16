#!/usr/bin/env python3


import pandas as pd
import seaborn as sns
import glob
import matplotlib.pyplot as plt

def main():
    path = "./measurements"
    files = glob.glob(path + "/*.csv")

    # combining measurements from different clients into one data frame
    list = []
    for f in files:
        df = pd.read_csv(f, header=0)
        list.append(df)
    df = pd.concat(list)
    

    sns.set_style("whitegrid")
    sns.boxplot(y="time-to-certificate", 
                x="csr-size",
                data=df,
                showfliers=False,
                width=0.2,)

    # show plot
    plt.ylim(0)
    plt.show()

if __name__ == '__main__':
    main()
