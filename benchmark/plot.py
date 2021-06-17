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
    

    fourNodes = df.loc[df["num-nodes"] == 4]
    sns.set_style("whitegrid")
    bplot = sns.boxplot(y="time-to-certificate", 
                x="csr-size",
                data=fourNodes,
                showfliers=False,
                width=0.2,)
    plt.ylim(0)

    bplot.set(xlabel ="CSR size in Bytes", ylabel = "time-to-certificate in ms", title ='Comparing CSR size')
    plt.show()

    node_count_plot = sns.boxplot(y="time-to-certificate", 
            x="num-nodes",
            data=df,
            showfliers=False,
            width=0.2,)
    plt.show()

if __name__ == '__main__':
    main()
