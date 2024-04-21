interface Window {
    preferRelax?: 0 | 1 | 2; // 0: vn, 1: relax, 2: autopilot
    favouriteMode?: number;
    graphType?: "rank" | "pp";
    userID?: number;
    currentUserID?: number;
    actualID?: number;
    graphPoints?: any[];
    countryRankPoints?: any[];
    graphName?: string;
    graphColor?: string;
    graphLabels?: string[];
    chart?: ApexCharts;
}

interface JQuery {
    modal: (_: string) => any;
    timeago: () => any;
    popup: (_: object) => any;
}
