package config

import (
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/scenario"
)

var structureJSON = []byte(`{
  "meta": {
    "title": "Ctrl00_allObj-Modified(1)(1)",
    "guid": "76b3f8c0-ed73-4732-b9de-1b1397477052"
  },
  "objects": {
    "02624185-cc26-4e98-92a1-d2008c36fd85": {
      "id": "02624185-cc26-4e98-92a1-d2008c36fd85",
      "type": "masterobject",
      "title": "BoxplotM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "Box start - 1.5 IQR",
          "def": "(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 ) - ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )) * 1.5))"
        },
        {
          "label": "First quartile",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )"
        },
        {
          "label": "Median",
          "def": "Median( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] )  )"
        },
        {
          "label": "Third quartile",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 )"
        },
        {
          "label": "Box end + 1.5 IQR",
          "def": "(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) + ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )) * 1.5))"
        }
      ],
      "visualization": "boxplot"
    },
    "1074afba-0a68-4ef3-b43a-02e500cd2fd8": {
      "id": "1074afba-0a68-4ef3-b43a-02e500cd2fd8",
      "type": "table",
      "title": "TableM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "extendsId": "LczZG",
      "visualization": "table"
    },
    "27908a64-d573-4aaa-ac04-b47932ba7995": {
      "id": "27908a64-d573-4aaa-ac04-b47932ba7995",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "29ee51e0-f9bb-4a4b-8686-96dcd314727c": {
      "id": "29ee51e0-f9bb-4a4b-8686-96dcd314727c",
      "type": "barchart",
      "title": "BarM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "extendsId": "WKUFp",
      "visualization": "barchart"
    },
    "2b86d4e7-dff0-4252-9c3f-f67a82626d7d": {
      "id": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Dim2M",
            "description": "Master Dim2"
          },
          "defs": [
            "Dim2"
          ],
          "labels": [
            "Dim2M"
          ]
        }
      ]
    },
    "3404f780-d272-4620-80b6-721e010f9152": {
      "id": "3404f780-d272-4620-80b6-721e010f9152",
      "type": "sheet",
      "title": "sheet2D",
      "children": {
        "8b7da3fa-45ee-4920-bd9d-419950be8f2e": "kpi",
        "a19b72ab-2ca8-4b69-8082-2c100aba8c3c": "pivot-table",
        "c66245be-8472-4900-9f05-e07a0141ed99": "scatterplot",
        "e94a7e71-0ebb-47c0-aeea-bd4b9eb8c69d": "text-image",
        "fa4229b9-6e9b-4bea-b516-c9c717167a29": "gauge"
      },
      "selectable": false
    },
    "35585276-91fd-45f2-8594-e28feae7a8cd": {
      "id": "35585276-91fd-45f2-8594-e28feae7a8cd",
      "type": "masterobject",
      "title": "DistributionM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(TransID)"
        }
      ],
      "visualization": "distributionplot"
    },
    "39978e68-e79b-4ea9-9b47-d4e26db4d155": {
      "id": "39978e68-e79b-4ea9-9b47-d4e26db4d155",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression2"
          ],
          "labels": [
            "Expression2"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "41dbb01c-d1bd-4528-be05-910ee565988b": {
      "id": "41dbb01c-d1bd-4528-be05-910ee565988b",
      "type": "sheet",
      "title": "sheet1D",
      "children": {
        "1074afba-0a68-4ef3-b43a-02e500cd2fd8": "table",
        "29ee51e0-f9bb-4a4b-8686-96dcd314727c": "barchart",
        "532bbe7a-8001-4b94-8b9e-441cc1e8aa48": "piechart",
        "5932e9a5-14e2-4f6d-8654-c719c288743e": "combochart",
        "7e3f0dd0-b5f2-4179-9172-d48011e167e9": "filterpane",
        "88a8c69b-5715-4c7c-b4ae-293da7a072d9": "treemap",
        "fcaf1755-a56b-4d33-a3b4-cebe86254e61": "linechart"
      },
      "selectable": false
    },
    "4e0bcb7f-8958-4d2e-b3e7-55f3509b2752": {
      "id": "4e0bcb7f-8958-4d2e-b3e7-55f3509b2752",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Sum(Exp1)"
          },
          "defs": [
            "=Sum(Expression1)"
          ],
          "labels": [
            "Sum(Exp1)"
          ]
        }
      ]
    },
    "532bbe7a-8001-4b94-8b9e-441cc1e8aa48": {
      "id": "532bbe7a-8001-4b94-8b9e-441cc1e8aa48",
      "type": "piechart",
      "title": "PieM",
      "description": "Master of Pie",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        }
      ],
      "extendsId": "bTLJu",
      "visualization": "piechart"
    },
    "53d3f92c-90e5-46a5-ab0f-019e9cf955ec": {
      "id": "53d3f92c-90e5-46a5-ab0f-019e9cf955ec",
      "type": "sheet",
      "title": "Accumulated and Mekko Charts",
      "children": {
        "KbVvkqF": "barchart",
        "TmzQd": "linechart",
        "kPpA": "mekkochart",
        "nPdzcpE": "combochart"
      },
      "selectable": false
    },
    "5932e9a5-14e2-4f6d-8654-c719c288743e": {
      "id": "5932e9a5-14e2-4f6d-8654-c719c288743e",
      "type": "combochart",
      "title": "ComboM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "extendsId": "QquqnR",
      "visualization": "combochart"
    },
    "6b9169cc-d6de-4fae-9520-d674e2ded148": {
      "id": "6b9169cc-d6de-4fae-9520-d674e2ded148",
      "type": "sheet",
      "title": "Histograms",
      "children": {
        "BvCKL": "histogram",
        "mqKjf": "histogram",
        "pHjJnUa": "histogram",
        "ykjZ": "histogram"
      },
      "selectable": false
    },
    "6bd32eab-6ae7-4b80-b3d3-e02f22e1d5b8": {
      "id": "6bd32eab-6ae7-4b80-b3d3-e02f22e1d5b8",
      "type": "masterobject",
      "title": "HistogramM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class([TransID],1100)"
          ]
        }
      ],
      "measures": [
        {
          "def": "Count([TransID])"
        }
      ],
      "visualization": "histogram"
    },
    "70256cd8-abe1-42b2-99e6-36665130ec03": {
      "id": "70256cd8-abe1-42b2-99e6-36665130ec03",
      "type": "sheet",
      "title": "Distribution Plots",
      "children": {
        "JPzUJ": "distributionplot",
        "QjPQnKP": "distributionplot",
        "sDuuGK": "distributionplot"
      },
      "selectable": false
    },
    "75d45738-aee0-4979-a393-e4f5daa6a509": {
      "id": "75d45738-aee0-4979-a393-e4f5daa6a509",
      "type": "sheet",
      "title": "Mekko Chart",
      "children": {
        "ynZFW": "mekkochart"
      },
      "selectable": false
    },
    "7e3f0dd0-b5f2-4179-9172-d48011e167e9": {
      "id": "7e3f0dd0-b5f2-4179-9172-d48011e167e9",
      "type": "filterpane",
      "title": "FilterPaneM",
      "children": {
        "27908a64-d573-4aaa-ac04-b47932ba7995": "listbox",
        "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": "listbox",
        "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": "listbox"
      },
      "selectable": false,
      "extendsId": "UmDGVm",
      "visualization": "filterpane"
    },
    "88a8c69b-5715-4c7c-b4ae-293da7a072d9": {
      "id": "88a8c69b-5715-4c7c-b4ae-293da7a072d9",
      "type": "treemap",
      "title": "TreeM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "extendsId": "JmdbDg",
      "visualization": "treemap"
    },
    "8b7da3fa-45ee-4920-bd9d-419950be8f2e": {
      "id": "8b7da3fa-45ee-4920-bd9d-419950be8f2e",
      "type": "kpi",
      "selectable": false,
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "kpi"
    },
    "9d8b2a78-4143-466c-8a5b-ffbf5fd54cba": {
      "id": "9d8b2a78-4143-466c-8a5b-ffbf5fd54cba",
      "type": "bookmark",
      "selectable": false
    },
    "ALGFsZ": {
      "id": "ALGFsZ",
      "type": "filterpane",
      "children": {
        "Fpwhwuy": "listbox",
        "QqCpG": "listbox",
        "gaB": "listbox",
        "kEhhTmG": "listbox",
        "kWcmhVt": "listbox",
        "kjHUmAD": "listbox",
        "krgrpm": "listbox",
        "mePVPA": "listbox",
        "wZrAD": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "Ahpt": {
      "id": "Ahpt",
      "type": "appprops",
      "selectable": false
    },
    "AjLSTb": {
      "id": "AjLSTb",
      "type": "masterobject",
      "title": "TextM",
      "selectable": false,
      "visualization": "text-image"
    },
    "BBabPm": {
      "id": "BBabPm",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(AsciiNum)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "combochart"
    },
    "BcJyGT": {
      "id": "BcJyGT",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "BhJBj": {
      "id": "BhJBj",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "BvCKL": {
      "id": "BvCKL",
      "type": "histogram",
      "title": "HistogramM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class([TransID],1100)"
          ]
        }
      ],
      "measures": [
        {
          "def": "Count([TransID])"
        }
      ],
      "extendsId": "6bd32eab-6ae7-4b80-b3d3-e02f22e1d5b8",
      "visualization": "histogram"
    },
    "CPKEjQP": {
      "id": "CPKEjQP",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(AsciiNum)"
        }
      ],
      "visualization": "barchart"
    },
    "CPPCtj": {
      "id": "CPPCtj",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        },
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Count(Dim3)"
        }
      ],
      "visualization": "combochart"
    },
    "CWHcvpF": {
      "id": "CWHcvpF",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression1)"
        },
        {
          "def": "Avg(Dim3)"
        }
      ],
      "visualization": "scatterplot"
    },
    "CnXmja": {
      "id": "CnXmja",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Count(Expression3)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "scatterplot"
    },
    "Dqtpwy": {
      "id": "Dqtpwy",
      "type": "filterpane",
      "children": {
        "39978e68-e79b-4ea9-9b47-d4e26db4d155": "listbox",
        "dcf57d11-a5f9-40cd-96a7-9e8cb737d6c9": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "DuJDJ": {
      "id": "DuJDJ",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "barchart"
    },
    "DuvEE": {
      "id": "DuvEE",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Dim1M",
            "description": "MasterDim1"
          },
          "defs": [
            "Dim1"
          ],
          "labels": [
            "Dim1M"
          ]
        }
      ]
    },
    "EFUmpQ": {
      "id": "EFUmpQ",
      "type": "table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Expression2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Num)"
        }
      ],
      "visualization": "table"
    },
    "EKSBrUL": {
      "id": "EKSBrUL",
      "type": "pivot-table",
      "title": "PivotM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "extendsId": "ZDxxg",
      "visualization": "pivot-table"
    },
    "ELvcvsJ": {
      "id": "ELvcvsJ",
      "type": "linechart",
      "title": "LineM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression2)"
        }
      ],
      "extendsId": "tEwrF",
      "visualization": "linechart"
    },
    "EaUj": {
      "id": "EaUj",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "Fpwhwuy": {
      "id": "Fpwhwuy",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression1"
          ],
          "labels": [
            "Expression1"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "FvFMQjL": {
      "id": "FvFMQjL",
      "type": "filterpane",
      "children": {
        "bapZzrB": "listbox",
        "maVjt": "listbox",
        "mtjmewY": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "GCuLj": {
      "id": "GCuLj",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(AsciiNum)"
        }
      ],
      "visualization": "piechart"
    },
    "GSNB": {
      "id": "GSNB",
      "type": "sheet",
      "title": "Sheet Calculation Condition",
      "children": {
        "DuJDJ": "barchart",
        "frSUmz": "text-image",
        "nuCHz": "filterpane"
      },
      "selectable": false
    },
    "GwyUc": {
      "id": "GwyUc",
      "type": "filterpane",
      "children": {
        "PutJZrX": "listbox",
        "QKPPSD": "listbox",
        "YTxPZp": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "HYwCuY": {
      "id": "HYwCuY",
      "type": "auto-chart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        },
        {
          "def": "Avg(Expression2)"
        }
      ],
      "visualization": "scatterplot"
    },
    "HkVwxK": {
      "id": "HkVwxK",
      "type": "text-image",
      "title": "TextM",
      "selectable": false,
      "extendsId": "AjLSTb",
      "visualization": "text-image"
    },
    "JMSnZmr": {
      "id": "JMSnZmr",
      "type": "waterfallchart",
      "selectable": false,
      "measures": [
        {
          "libraryId": "gWKjj"
        },
        {
          "libraryId": "uSFruQ"
        },
        {
          "def": "Avg(AsciiAlpha)"
        }
      ],
      "visualization": "waterfallchart"
    },
    "JPzUJ": {
      "id": "JPzUJ",
      "type": "distributionplot",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        },
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "distributionplot"
    },
    "JhmJf": {
      "id": "JhmJf",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "JmdbDg": {
      "id": "JmdbDg",
      "type": "masterobject",
      "title": "TreeM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "visualization": "treemap"
    },
    "JpcpmT": {
      "id": "JpcpmT",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "Jsamsb": {
      "id": "Jsamsb",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "JuhBpt": {
      "id": "JuhBpt",
      "type": "sheet",
      "title": "MasterSheet1",
      "children": {
        "EKSBrUL": "pivot-table",
        "ELvcvsJ": "linechart",
        "EaUj": "map",
        "HkVwxK": "text-image",
        "NrnEY": "piechart",
        "PUvSAN": "scatterplot",
        "QAr": "table",
        "SyZnb": "gauge",
        "VevpJm": "treemap",
        "XjwmbU": "barchart",
        "bZCvHFB": "combochart",
        "naSGALJ": "filterpane",
        "ppUMK": "kpi"
      },
      "selectable": false
    },
    "KbVvkqF": {
      "id": "KbVvkqF",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "label": "Avg(Expression3)",
          "def": "RangeSum(Above(Avg(Expression3) + Sum({1} 0), 0, 6))"
        },
        {
          "label": "Avg(Expression1)",
          "def": "RangeSum(Above(Avg(Expression1) + Sum({1} 0), 0, 6))"
        },
        {
          "label": "Avg(Expression2)",
          "def": "RangeSum(Above(Avg(Expression2) + Sum({1} 0), 0, 6))"
        }
      ],
      "visualization": "barchart"
    },
    "KtpdU": {
      "id": "KtpdU",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression2)"
        }
      ],
      "visualization": "linechart"
    },
    "KuaUFm": {
      "id": "KuaUFm",
      "type": "auto-chart",
      "selectable": true,
      "dimensions": [
        {
          "labelExpression": "='Dim label with expr' + Count(Dim1)",
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "LNabD": {
      "id": "LNabD",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        },
        {
          "def": "Avg(Expression2)"
        },
        {
          "def": "Avg(Num)"
        }
      ],
      "visualization": "combochart"
    },
    "LQCanwJ": {
      "id": "LQCanwJ",
      "type": "table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "table"
    },
    "LXJqzzq": {
      "id": "LXJqzzq",
      "type": "treemap",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(TransLineID)"
        }
      ],
      "visualization": "treemap"
    },
    "LczZG": {
      "id": "LczZG",
      "type": "masterobject",
      "title": "TableM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "table"
    },
    "LkpcR": {
      "id": "LkpcR",
      "type": "table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        },
        {
          "def": "Count(Alpha)"
        },
        {
          "def": "Count(Dim3)"
        }
      ],
      "visualization": "table"
    },
    "LoadModel": {
      "id": "LoadModel",
      "type": "LoadModel",
      "selectable": false
    },
    "MJcFdz": {
      "id": "MJcFdz",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "NBMVsP": {
      "id": "NBMVsP",
      "type": "treemap",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "visualization": "treemap"
    },
    "NrnEY": {
      "id": "NrnEY",
      "type": "piechart",
      "title": "PieM",
      "description": "Master of Pie",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        }
      ],
      "extendsId": "bTLJu",
      "visualization": "piechart"
    },
    "NxTZDm": {
      "id": "NxTZDm",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        }
      ],
      "visualization": "piechart"
    },
    "PUvSAN": {
      "id": "PUvSAN",
      "type": "scatterplot",
      "title": "ScatterM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Count(Expression3)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "extendsId": "VqPXn",
      "visualization": "scatterplot"
    },
    "PaMnJfP": {
      "id": "PaMnJfP",
      "type": "gauge",
      "selectable": false,
      "measures": [
        {
          "def": "Avg(Expression2)"
        }
      ],
      "visualization": "gauge"
    },
    "PhZtQ": {
      "id": "PhZtQ",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(TransID)"
        },
        {
          "def": "Count(Num)"
        },
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "scatterplot"
    },
    "PutJZrX": {
      "id": "PutJZrX",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            "AsciiAlpha"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "QAr": {
      "id": "QAr",
      "type": "table",
      "title": "TableM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "extendsId": "LczZG",
      "visualization": "table"
    },
    "QKPPSD": {
      "id": "QKPPSD",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            "Num"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "QjPQnKP": {
      "id": "QjPQnKP",
      "type": "distributionplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(TransID)"
        }
      ],
      "visualization": "distributionplot"
    },
    "Qkpy": {
      "id": "Qkpy",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        },
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        }
      ],
      "visualization": "barchart"
    },
    "QqCpG": {
      "id": "QqCpG",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            "Num"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "QqFVRmp": {
      "id": "QqFVRmp",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Property"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "Value",
          "def": "Sum(Property.Value)"
        }
      ],
      "visualization": "map"
    },
    "QquqnR": {
      "id": "QquqnR",
      "type": "masterobject",
      "title": "ComboM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "combochart"
    },
    "RaQyNmU": {
      "id": "RaQyNmU",
      "type": "treemap",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression1)"
        }
      ],
      "visualization": "treemap"
    },
    "RednFCk": {
      "id": "RednFCk",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "libraryId": "pmAQd"
        }
      ],
      "visualization": "barchart"
    },
    "RmhqnQT": {
      "id": "RmhqnQT",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        },
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "scatterplot"
    },
    "RznsJc": {
      "id": "RznsJc",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(TransLineID)"
        }
      ],
      "visualization": "barchart"
    },
    "SumfyV": {
      "id": "SumfyV",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "SyZnb": {
      "id": "SyZnb",
      "type": "gauge",
      "title": "GuaugeM",
      "selectable": false,
      "measures": [
        {
          "def": "Avg(Expression2)"
        }
      ],
      "extendsId": "bMafQ",
      "visualization": "gauge"
    },
    "TMAvcJs": {
      "id": "TMAvcJs",
      "type": "filterpane",
      "children": {
        "qBXxP": "listbox",
        "wXKASnP": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "TQfAas": {
      "id": "TQfAas",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        },
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "barchart"
    },
    "TfJeGV": {
      "id": "TfJeGV",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "visualization": "barchart"
    },
    "TmfVzX": {
      "id": "TmfVzX",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "piechart"
    },
    "TmzQd": {
      "id": "TmzQd",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "Count(Expression2)",
          "def": "RangeSum(Above(Count(Expression2) + Sum({1} 0), 0, 6))"
        }
      ],
      "visualization": "linechart"
    },
    "UVahMU": {
      "id": "UVahMU",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        }
      ],
      "visualization": "piechart"
    },
    "Ubyp": {
      "id": "Ubyp",
      "type": "text-image",
      "selectable": false,
      "measures": [
        {
          "def": "'AsciiAlpha Counts: ' \u0026 Count(AsciiAlpha) \u0026chr(13)\u0026 'AsciiNum Counts: ' \u0026 Count(AsciiNum) \u0026chr(13)\u0026 'Num Counts: ' \u0026 Count(Num)"
        }
      ],
      "visualization": "text-image"
    },
    "UmDGVm": {
      "id": "UmDGVm",
      "type": "masterobject",
      "title": "FilterPaneM",
      "children": {
        "27908a64-d573-4aaa-ac04-b47932ba7995": "listbox",
        "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": "listbox",
        "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "VPmt": {
      "id": "VPmt",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "piechart"
    },
    "VQTSUM": {
      "id": "VQTSUM",
      "type": "sheet",
      "title": "Filter Pane  and Text \u0026 Image",
      "children": {
        "ALGFsZ": "filterpane",
        "GwyUc": "filterpane",
        "Jsamsb": "text-image",
        "SumfyV": "text-image",
        "TMAvcJs": "filterpane",
        "Ubyp": "text-image",
        "VvEUTD": "text-image",
        "ZZmPP": "text-image",
        "ZsbBur": "text-image",
        "hJUssV": "text-image",
        "jpTaBxJ": "text-image"
      },
      "selectable": false
    },
    "VeQpVnP": {
      "id": "VeQpVnP",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiAlpha)"
        },
        {
          "def": "Count(Dim2)"
        },
        {
          "def": "Count(Dim1)"
        },
        {
          "def": "Count(Dim3)"
        },
        {
          "def": "Count(Alpha)"
        },
        {
          "def": "Avg(AsciiNum)"
        }
      ],
      "visualization": "linechart"
    },
    "VevpJm": {
      "id": "VevpJm",
      "type": "treemap",
      "title": "TreeM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "extendsId": "JmdbDg",
      "visualization": "treemap"
    },
    "VqPXn": {
      "id": "VqPXn",
      "type": "masterobject",
      "title": "ScatterM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Count(Expression3)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "scatterplot"
    },
    "VvEUTD": {
      "id": "VvEUTD",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "WKUFp": {
      "id": "WKUFp",
      "type": "masterobject",
      "title": "BarM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "WNmKY": {
      "id": "WNmKY",
      "type": "auto-chart",
      "selectable": true,
      "dimensions": [
        {
          "labelExpression": "='Dim label with set expression'+Avg({\u003cDim1={'A'}\u003e} Dim1)",
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            "Normal Label for Alpha"
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "WYjhNJ": {
      "id": "WYjhNJ",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(TransID)"
        },
        {
          "libraryId": "uSFruQ"
        }
      ],
      "visualization": "piechart"
    },
    "XAPJvj": {
      "id": "XAPJvj",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        },
        {
          "def": "Avg(Expression1)"
        },
        {
          "def": "Avg(Expression2)"
        }
      ],
      "visualization": "barchart"
    },
    "XAhMkS": {
      "id": "XAhMkS",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "def": "Avg(Num)"
        },
        {
          "def": "Avg(AsciiNum)"
        },
        {
          "def": "Count(AsciiAlpha)"
        },
        {
          "def": "Count(AsciiNum)"
        }
      ],
      "visualization": "combochart"
    },
    "XbVzTCh": {
      "id": "XbVzTCh",
      "type": "pivot-table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "=Num"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "=Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression2)"
        }
      ],
      "visualization": "pivot-table"
    },
    "XffY": {
      "id": "XffY",
      "type": "filterpane",
      "children": {
        "ggVmAvr": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "XjwmbU": {
      "id": "XjwmbU",
      "type": "barchart",
      "title": "BarM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "extendsId": "WKUFp",
      "visualization": "barchart"
    },
    "XwGwH": {
      "id": "XwGwH",
      "type": "auto-chart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "table"
    },
    "YTxPZp": {
      "id": "YTxPZp",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "ZCgtkh": {
      "id": "ZCgtkh",
      "type": "filterpane",
      "children": {
        "vMmpj": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "ZDxxg": {
      "id": "ZDxxg",
      "type": "masterobject",
      "title": "PivotM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "pivot-table"
    },
    "ZRybA": {
      "id": "ZRybA",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "linechart"
    },
    "ZZmPP": {
      "id": "ZZmPP",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "ZsbBur": {
      "id": "ZsbBur",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "a19b72ab-2ca8-4b69-8082-2c100aba8c3c": {
      "id": "a19b72ab-2ca8-4b69-8082-2c100aba8c3c",
      "type": "pivot-table",
      "title": "PivotM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "extendsId": "ZDxxg",
      "visualization": "pivot-table"
    },
    "aFjjST": {
      "id": "aFjjST",
      "type": "sheet",
      "title": "Maps",
      "children": {
        "MJcFdz": "map",
        "QqFVRmp": "map",
        "pCYhUmk": "map",
        "teTUAr": "map",
        "yvyZVN": "map"
      },
      "selectable": false
    },
    "aGvPJp": {
      "id": "aGvPJp",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(TransLineID)"
        }
      ],
      "visualization": "linechart"
    },
    "ab563802-8c69-427e-929e-0966bdaf07eb": {
      "id": "ab563802-8c69-427e-929e-0966bdaf07eb",
      "type": "sheet",
      "title": "Rose chart",
      "children": {
        "WYjhNJ": "piechart",
        "dxre": "piechart"
      },
      "selectable": false
    },
    "anmpmjq": {
      "id": "anmpmjq",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression1)"
        },
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "barchart"
    },
    "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": {
      "id": "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            "Alpha"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "bKn": {
      "id": "bKn",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression2)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "scatterplot"
    },
    "bMafQ": {
      "id": "bMafQ",
      "type": "masterobject",
      "title": "GuaugeM",
      "selectable": false,
      "measures": [
        {
          "def": "Avg(Expression2)"
        }
      ],
      "visualization": "gauge"
    },
    "bTLJu": {
      "id": "bTLJu",
      "type": "masterobject",
      "title": "PieM",
      "description": "Master of Pie",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "piechart"
    },
    "bZCvHFB": {
      "id": "bZCvHFB",
      "type": "combochart",
      "title": "ComboM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "extendsId": "QquqnR",
      "visualization": "combochart"
    },
    "bapZzrB": {
      "id": "bapZzrB",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "c66245be-8472-4900-9f05-e07a0141ed99": {
      "id": "c66245be-8472-4900-9f05-e07a0141ed99",
      "type": "scatterplot",
      "title": "ScatterM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Count(Expression3)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "extendsId": "VqPXn",
      "visualization": "scatterplot"
    },
    "cQTurT": {
      "id": "cQTurT",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "libraryId": "gWKjj"
        }
      ],
      "visualization": "barchart"
    },
    "cbVHSUk": {
      "id": "cbVHSUk",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression2)"
        },
        {
          "def": "Avg(Num)"
        },
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "linechart"
    },
    "cf5e4644-fd93-4b01-848b-5ba69124987b": {
      "id": "cf5e4644-fd93-4b01-848b-5ba69124987b",
      "type": "bookmark",
      "selectable": false
    },
    "cwqJNQ": {
      "id": "cwqJNQ",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33": {
      "id": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Dims123D",
            "description": "DrillDown Dims"
          },
          "defs": [
            "Dim1",
            "Dim2",
            "Dim3"
          ]
        }
      ]
    },
    "dYZkN": {
      "id": "dYZkN",
      "type": "sheet",
      "title": "Tables \u0026 Pivot Tables",
      "children": {
        "EFUmpQ": "table",
        "LkpcR": "table",
        "XbVzTCh": "pivot-table",
        "pAhvxJQ": "pivot-table",
        "uHzBpZv": "table",
        "wGKBMM": "pivot-table"
      },
      "selectable": false
    },
    "dcf57d11-a5f9-40cd-96a7-9e8cb737d6c9": {
      "id": "dcf57d11-a5f9-40cd-96a7-9e8cb737d6c9",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression1"
          ],
          "labels": [
            "Expression1"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "dfbgtd": {
      "id": "dfbgtd",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        }
      ],
      "visualization": "barchart"
    },
    "dmjZp": {
      "id": "dmjZp",
      "type": "masterobject",
      "title": "KPIM",
      "selectable": false,
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "visualization": "kpi"
    },
    "dxre": {
      "id": "dxre",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(TransID)"
        },
        {
          "libraryId": "uSFruQ"
        }
      ],
      "visualization": "piechart"
    },
    "e097e3c4-8304-495b-be67-482a2eba4fd1": {
      "id": "e097e3c4-8304-495b-be67-482a2eba4fd1",
      "type": "sheet",
      "title": "Org chart",
      "children": {
        "vHznrgh": "sn-org-chart"
      },
      "selectable": false
    },
    "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": {
      "id": "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "e85e2002-3daf-44b6-a052-784da3323444": {
      "id": "e85e2002-3daf-44b6-a052-784da3323444",
      "type": "sheet",
      "title": "more maps",
      "children": {
        "JhmJf": "map",
        "hjvTwJp": "map"
      },
      "selectable": false
    },
    "e94a7e71-0ebb-47c0-aeea-bd4b9eb8c69d": {
      "id": "e94a7e71-0ebb-47c0-aeea-bd4b9eb8c69d",
      "type": "text-image",
      "title": "TextM",
      "selectable": false,
      "extendsId": "AjLSTb",
      "visualization": "text-image"
    },
    "e98010b2-a77e-4c5c-9612-1c3329713141": {
      "id": "e98010b2-a77e-4c5c-9612-1c3329713141",
      "type": "sheet",
      "title": "Waterfall",
      "children": {
        "Dqtpwy": "filterpane",
        "JMSnZmr": "waterfallchart",
        "vTHMcf": "waterfallchart"
      },
      "selectable": false
    },
    "eXPkSg": {
      "id": "eXPkSg",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "combochart"
    },
    "efcNyfp": {
      "id": "efcNyfp",
      "type": "linechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(AsciiNum)"
        }
      ],
      "visualization": "linechart"
    },
    "esXkULt": {
      "id": "esXkULt",
      "type": "sheet",
      "title": "MasterSheet2",
      "children": {
        "RednFCk": "barchart",
        "XffY": "filterpane",
        "ZCgtkh": "filterpane",
        "cQTurT": "barchart",
        "mGp": "barchart",
        "tasxmrc": "filterpane"
      },
      "selectable": false
    },
    "evzqxe": {
      "id": "evzqxe",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Num)"
        },
        {
          "def": "Count(TransLineID)"
        }
      ],
      "visualization": "scatterplot"
    },
    "f2a50cb3-a7e1-40ac-a015-bc4378773312": {
      "id": "f2a50cb3-a7e1-40ac-a015-bc4378773312",
      "type": "sheet",
      "title": "Auto Charts",
      "children": {
        "HYwCuY": "auto-chart",
        "KuaUFm": "auto-chart",
        "WNmKY": "auto-chart",
        "XwGwH": "auto-chart",
        "jCjQxvm": "auto-chart"
      },
      "selectable": false
    },
    "f57d852f-f248-4a1a-a15b-015e60147cf8": {
      "id": "f57d852f-f248-4a1a-a15b-015e60147cf8",
      "type": "sheet",
      "title": "Boxplots",
      "children": {
        "fCHrRdn": "boxplot",
        "uaHYE": "boxplot"
      },
      "selectable": false
    },
    "fCHrRdn": {
      "id": "fCHrRdn",
      "type": "boxplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "First whisker",
          "def": "Rangemax(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.25 ) - ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.25 )) * 1.5), Min( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] )  ))"
        },
        {
          "label": "Box start",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.25 )"
        },
        {
          "label": "Center line",
          "def": "Median( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] )  )"
        },
        {
          "label": "Box end",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.75 )"
        },
        {
          "label": "Last whisker",
          "def": "Rangemin(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.75 ) + ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) ,0.25 )) * 1.5), Max( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] )  ))"
        }
      ],
      "visualization": "boxplot"
    },
    "fEPFmQ": {
      "id": "fEPFmQ",
      "type": "sheet",
      "title": "Scatter Plots",
      "children": {
        "CWHcvpF": "scatterplot",
        "PhZtQ": "scatterplot",
        "RmhqnQT": "scatterplot",
        "bKn": "scatterplot",
        "evzqxe": "scatterplot",
        "pYkJH": "scatterplot",
        "zJSbv": "scatterplot"
      },
      "selectable": false
    },
    "fJKJZc": {
      "id": "fJKJZc",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(TransID)"
        }
      ],
      "visualization": "piechart"
    },
    "fLXy": {
      "id": "fLXy",
      "type": "kpi",
      "title": "KPIM",
      "selectable": false,
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "extendsId": "dmjZp",
      "visualization": "kpi"
    },
    "fa4229b9-6e9b-4bea-b516-c9c717167a29": {
      "id": "fa4229b9-6e9b-4bea-b516-c9c717167a29",
      "type": "gauge",
      "title": "GuaugeM",
      "selectable": false,
      "measures": [
        {
          "def": "Avg(Expression2)"
        }
      ],
      "extendsId": "bMafQ",
      "visualization": "gauge"
    },
    "fcaf1755-a56b-4d33-a3b4-cebe86254e61": {
      "id": "fcaf1755-a56b-4d33-a3b4-cebe86254e61",
      "type": "linechart",
      "title": "LineM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression2)"
        }
      ],
      "extendsId": "tEwrF",
      "visualization": "linechart"
    },
    "fjETFn": {
      "id": "fjETFn",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        },
        {
          "def": "Count(Alpha)"
        },
        {
          "libraryId": "pmAQd"
        }
      ],
      "visualization": "combochart"
    },
    "frSUmz": {
      "id": "frSUmz",
      "type": "text-image",
      "selectable": false,
      "measures": [
        {
          "def": "//FieldValue('Dim1',GetSelectedCount(Dim1))\n//if(RANK('Dim1')=1,'N/A',FieldValue('Dim1',1))\n//FieldValue('Dim1',2)\n//IF(GetSelectedCount('Dim1')=1,'N/A','One selected')\n//GetSelectedCount('Dim1')\n//Rank('Dim')\n//GetCurrentSelections('Dim1')\nGetSelectedCount(Dim1)\n"
        },
        {
          "def": "//FieldValue('Dim1',GetSelectedCount(Dim1))\n//if(RANK('Dim1')=1,'N/A',FieldValue('Dim1',1))\n//FieldValue('Dim1',2)\n//IF(GetSelectedCount('Dim1')=1,'N/A','One selected')\n//GetSelectedCount('Dim1')\n//Rank('Dim')\nGetCurrentSelections('Dim1')\n//GetSelectedCount(Dim1)\n"
        }
      ],
      "visualization": "text-image"
    },
    "gJfAjj": {
      "id": "gJfAjj",
      "type": "sheet",
      "title": "Bar Charts",
      "children": {
        "BcJyGT": "barchart",
        "BhJBj": "barchart",
        "CPKEjQP": "barchart",
        "Qkpy": "barchart",
        "RznsJc": "barchart",
        "TQfAas": "barchart",
        "TfJeGV": "barchart",
        "XAPJvj": "barchart",
        "anmpmjq": "barchart",
        "dfbgtd": "barchart",
        "gNM": "barchart",
        "hhaN": "combochart",
        "jEPxKRr": "barchart",
        "mSNXpjj": "barchart",
        "pcmJZa": "barchart",
        "tmnqgR": "barchart"
      },
      "selectable": false
    },
    "gNM": {
      "id": "gNM",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression3)"
        },
        {
          "def": "Avg(Expression2)"
        },
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "gWKjj": {
      "id": "gWKjj",
      "type": "measure",
      "selectable": false,
      "measures": [
        {
          "meta": {
            "title": "Avg(Exp2)"
          },
          "label": "Avg(Exp2)",
          "def": "Avg(Expression2)"
        }
      ]
    },
    "gaB": {
      "id": "gaB",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "visualization": "listbox"
    },
    "gdJaz": {
      "id": "gdJaz",
      "type": "treemap",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        },
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        },
        {
          "defs": [
            "Dim3"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Avg(TransID)"
        }
      ],
      "visualization": "treemap"
    },
    "ggVmAvr": {
      "id": "ggVmAvr",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        }
      ],
      "visualization": "listbox"
    },
    "hJUssV": {
      "id": "hJUssV",
      "type": "text-image",
      "selectable": false,
      "measures": [
        {
          "def": "'Num Counts: ' \u0026 Count(Num) \u0026chr(13)\u0026 'Alpha: ' \u0026 Count(Alpha)"
        }
      ],
      "visualization": "text-image"
    },
    "hQDEATt": {
      "id": "hQDEATt",
      "type": "sheet",
      "title": "TreeMaps",
      "children": {
        "LXJqzzq": "treemap",
        "RaQyNmU": "treemap",
        "gdJaz": "treemap"
      },
      "selectable": false
    },
    "hSTsZt": {
      "id": "hSTsZt",
      "type": "sheet",
      "title": "Line charts",
      "children": {
        "VeQpVnP": "linechart",
        "ZRybA": "linechart",
        "aGvPJp": "linechart",
        "cbVHSUk": "linechart",
        "efcNyfp": "linechart"
      },
      "selectable": false
    },
    "hhaN": {
      "id": "hhaN",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "def": "Avg(Expression1)"
        },
        {
          "def": "Avg(Expression2)"
        },
        {
          "def": "Avg(Expression3)"
        }
      ],
      "visualization": "combochart"
    },
    "hjvTwJp": {
      "id": "hjvTwJp",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "hkDzPW": {
      "id": "hkDzPW",
      "type": "pivot-table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "pivot-table"
    },
    "jCjQxvm": {
      "id": "jCjQxvm",
      "type": "auto-chart",
      "selectable": false,
      "measures": [
        {
          "def": "Avg(Expression2)"
        }
      ],
      "visualization": "kpi"
    },
    "jEPxKRr": {
      "id": "jEPxKRr",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(TransLineID)"
        }
      ],
      "visualization": "barchart"
    },
    "jpTaBxJ": {
      "id": "jpTaBxJ",
      "type": "text-image",
      "selectable": false,
      "measures": [
        {
          "def": "'Num Counts: ' \u0026 Count(Num) \u0026chr(13)\u0026 'TranLineID: ' \u0026 Count(TransLineID) \u0026chr(13)\u0026 'TranID: ' \u0026 Count(TransID) \u0026chr(13)\u0026 'Dim1M: ' \u0026 Count(Dim1) \u0026chr(13)\u0026 'Dim2M: ' \u0026 Count(Dim2) \u0026chr(13)\u0026 'Dim123D: ' \u0026 Count(Dim3) \u0026chr(13)\u0026 'Expression1: ' \u0026 Count(Expression1) \u0026chr(13)\u0026 'Expression2: ' \u0026 Count(Expression2) \u0026chr(13)\u0026 'Expression3: ' \u0026 Count(Expression3)"
        }
      ],
      "visualization": "text-image"
    },
    "kEhhTmG": {
      "id": "kEhhTmG",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            "TransID"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "kPpA": {
      "id": "kPpA",
      "type": "mekkochart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        },
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "libraryId": "gWKjj"
        }
      ],
      "visualization": "mekkochart"
    },
    "kWcmhVt": {
      "id": "kWcmhVt",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression3"
          ],
          "labels": [
            "Expression3"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "kjHUmAD": {
      "id": "kjHUmAD",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression2"
          ],
          "labels": [
            "Expression2"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "krgrpm": {
      "id": "krgrpm",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "visualization": "listbox"
    },
    "mBshXB": {
      "id": "mBshXB",
      "type": "sheet",
      "title": "sheet1",
      "children": {
        "FvFMQjL": "filterpane",
        "JpcpmT": "barchart",
        "KtpdU": "linechart",
        "LQCanwJ": "table",
        "NBMVsP": "treemap",
        "TmfVzX": "piechart",
        "tCxXew": "combochart"
      },
      "selectable": false
    },
    "mGp": {
      "id": "mGp",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "libraryId": "uSFruQ"
        }
      ],
      "visualization": "barchart"
    },
    "mSNXpjj": {
      "id": "mSNXpjj",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "AsciiNum"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "libraryId": "pmAQd"
        }
      ],
      "visualization": "barchart"
    },
    "maVjt": {
      "id": "maVjt",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            "Alpha"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "mePVPA": {
      "id": "mePVPA",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            "TransLineID"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "mkeka": {
      "id": "mkeka",
      "type": "sheet",
      "title": "Combo Charts",
      "children": {
        "BBabPm": "combochart",
        "CPPCtj": "combochart",
        "LNabD": "combochart",
        "XAhMkS": "combochart",
        "eXPkSg": "combochart",
        "fjETFn": "combochart"
      },
      "selectable": false
    },
    "mqHMtkm": {
      "id": "mqHMtkm",
      "type": "pivot-table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "pivot-table"
    },
    "mqKjf": {
      "id": "mqKjf",
      "type": "histogram",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class(aggr([Num],[Num]),5)"
          ],
          "labels": [
            "Num"
          ]
        }
      ],
      "measures": [
        {
          "label": "Frequency",
          "def": "Count([Num])"
        }
      ],
      "visualization": "histogram"
    },
    "mtjmewY": {
      "id": "mtjmewY",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "nPdzcpE": {
      "id": "nPdzcpE",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "measures": [
        {
          "label": "Avg(Num)",
          "def": "RangeSum(Above(Avg(Num) + Sum({1} 0), 0, 6))"
        },
        {
          "label": "Avg(AsciiNum)",
          "def": "RangeSum(Above(Avg(AsciiNum) + Sum({1} 0), 0, 6))"
        },
        {
          "label": "Count(AsciiAlpha)",
          "def": "RangeSum(Above(Count(AsciiAlpha) + Sum({1} 0), 0, 6))"
        },
        {
          "label": "Count(AsciiNum)",
          "def": "RangeSum(Above(Count(AsciiNum) + Sum({1} 0), 0, 6))"
        }
      ],
      "visualization": "combochart"
    },
    "naSGALJ": {
      "id": "naSGALJ",
      "type": "filterpane",
      "title": "FilterPaneM",
      "children": {
        "27908a64-d573-4aaa-ac04-b47932ba7995": "listbox",
        "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": "listbox",
        "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": "listbox"
      },
      "selectable": false,
      "extendsId": "UmDGVm",
      "visualization": "filterpane"
    },
    "nuCHz": {
      "id": "nuCHz",
      "type": "filterpane",
      "children": {
        "uQxPN": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "pAhvxJQ": {
      "id": "pAhvxJQ",
      "type": "pivot-table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "=Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Dim1)"
        }
      ],
      "visualization": "pivot-table"
    },
    "pCYhUmk": {
      "id": "pCYhUmk",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "pHjJnUa": {
      "id": "pHjJnUa",
      "type": "histogram",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class(aggr([Expression2],[Expression2]),0.575)"
          ],
          "labels": [
            "Expression2"
          ]
        }
      ],
      "measures": [
        {
          "label": "Frequency",
          "def": "Count([Expression2])"
        }
      ],
      "visualization": "histogram"
    },
    "pYkJH": {
      "id": "pYkJH",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Num)"
        },
        {
          "def": "Count(TransLineID)"
        }
      ],
      "visualization": "scatterplot"
    },
    "pcmJZa": {
      "id": "pcmJZa",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression1)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "barchart"
    },
    "pmAQd": {
      "id": "pmAQd",
      "type": "measure",
      "selectable": false,
      "measures": [
        {
          "meta": {
            "title": "Sum(Exp1)"
          },
          "label": "Sum(Exp1)",
          "def": "Sum(Expression1)"
        }
      ]
    },
    "ppUMK": {
      "id": "ppUMK",
      "type": "kpi",
      "title": "KPIM",
      "selectable": false,
      "measures": [
        {
          "def": "Sum(Expression1)"
        }
      ],
      "extendsId": "dmjZp",
      "visualization": "kpi"
    },
    "ppjDSJ": {
      "id": "ppjDSJ",
      "type": "sheet",
      "title": "sheet2",
      "children": {
        "CnXmja": "scatterplot",
        "PaMnJfP": "gauge",
        "cwqJNQ": "text-image",
        "fLXy": "kpi",
        "hkDzPW": "pivot-table",
        "mqHMtkm": "pivot-table"
      },
      "selectable": false
    },
    "qBXxP": {
      "id": "qBXxP",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            "Alpha"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "qjwe": {
      "id": "qjwe",
      "type": "masterobject",
      "title": "MApM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Public_Toilets.Name"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Public_Toilets.Point"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "rkuwZh": {
      "id": "rkuwZh",
      "type": "sheet",
      "title": "Pie chart's",
      "children": {
        "GCuLj": "piechart",
        "NxTZDm": "piechart",
        "UVahMU": "piechart",
        "VPmt": "piechart",
        "fJKJZc": "piechart",
        "uHfjq": "piechart"
      },
      "selectable": false
    },
    "sDuuGK": {
      "id": "sDuuGK",
      "type": "distributionplot",
      "title": "DistributionM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(TransID)"
        }
      ],
      "extendsId": "35585276-91fd-45f2-8594-e28feae7a8cd",
      "visualization": "distributionplot"
    },
    "tCxXew": {
      "id": "tCxXew",
      "type": "combochart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Expression2)"
        },
        {
          "def": "Sum(Expression3)"
        }
      ],
      "visualization": "combochart"
    },
    "tEwrF": {
      "id": "tEwrF",
      "type": "masterobject",
      "title": "LineM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Expression2)"
        }
      ],
      "visualization": "linechart"
    },
    "tasxmrc": {
      "id": "tasxmrc",
      "type": "filterpane",
      "children": {
        "vXhmc": "listbox"
      },
      "selectable": false,
      "visualization": "filterpane"
    },
    "teTUAr": {
      "id": "teTUAr",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "map"
    },
    "tmnqgR": {
      "id": "tmnqgR",
      "type": "barchart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "def": "Avg(Num)"
        },
        {
          "def": "Avg(Expression1)"
        }
      ],
      "visualization": "barchart"
    },
    "uHfjq": {
      "id": "uHfjq",
      "type": "piechart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Sum(Num)"
        }
      ],
      "visualization": "piechart"
    },
    "uHzBpZv": {
      "id": "uHzBpZv",
      "type": "table",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        },
        {
          "defs": [
            "Alpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        },
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "visualization": "table"
    },
    "uQxPN": {
      "id": "uQxPN",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Dim1"
          ],
          "labels": [
            "Dim1"
          ]
        }
      ],
      "visualization": "listbox"
    },
    "uSFruQ": {
      "id": "uSFruQ",
      "type": "measure",
      "selectable": false,
      "measures": [
        {
          "meta": {
            "title": "Max(Exp3)"
          },
          "label": "Max(Exp3)",
          "def": "Max(Expression3)"
        }
      ]
    },
    "uaHYE": {
      "id": "uaHYE",
      "type": "boxplot",
      "title": "BoxplotM",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "TransLineID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "Box start - 1.5 IQR",
          "def": "(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 ) - ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )) * 1.5))"
        },
        {
          "label": "First quartile",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )"
        },
        {
          "label": "Median",
          "def": "Median( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] )  )"
        },
        {
          "label": "Third quartile",
          "def": "Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 )"
        },
        {
          "label": "Box end + 1.5 IQR",
          "def": "(Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) + ((Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.75 ) - Fractile( total \u003c[TransLineID]\u003e Aggr( Avg(TransID), [TransLineID], [Alpha] ) , 0.25 )) * 1.5))"
        }
      ],
      "extendsId": "02624185-cc26-4e98-92a1-d2008c36fd85",
      "visualization": "boxplot"
    },
    "vHznrgh": {
      "id": "vHznrgh",
      "type": "sn-org-chart",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "EmployeeID"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "ManagerID"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(EmployeeID)"
        }
      ],
      "visualization": "sn-org-chart"
    },
    "vMmpj": {
      "id": "vMmpj",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "visualization": "listbox"
    },
    "vTHMcf": {
      "id": "vTHMcf",
      "type": "waterfallchart",
      "selectable": false,
      "measures": [
        {
          "def": "Count(Expression1)"
        },
        {
          "def": "Count(Expression2)"
        },
        {
          "def": "Count(Expression3)"
        },
        {
          "def": "Count(Expression3)"
        }
      ],
      "visualization": "waterfallchart"
    },
    "vXhmc": {
      "id": "vXhmc",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "visualization": "listbox"
    },
    "wGKBMM": {
      "id": "wGKBMM",
      "type": "pivot-table",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Expression1"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "=Expression2"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(Num)"
        },
        {
          "def": "Avg(Expression3)"
        },
        {
          "def": "Sum(AsciiNum)"
        }
      ],
      "visualization": "pivot-table"
    },
    "wXKASnP": {
      "id": "wXKASnP",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Num"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "visualization": "listbox"
    },
    "wZrAD": {
      "id": "wZrAD",
      "type": "listbox",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33"
        }
      ],
      "visualization": "listbox"
    },
    "ykjZ": {
      "id": "ykjZ",
      "type": "histogram",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class(aggr([Expression1],[Expression1]),60)"
          ],
          "labels": [
            "Expression1"
          ]
        }
      ],
      "measures": [
        {
          "label": "Frequency",
          "def": "Count([Expression1])"
        }
      ],
      "visualization": "histogram"
    },
    "ynZFW": {
      "id": "ynZFW",
      "type": "mekkochart",
      "selectable": true,
      "dimensions": [
        {
          "libraryId": "DuvEE"
        },
        {
          "libraryId": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d"
        }
      ],
      "measures": [
        {
          "libraryId": "gWKjj"
        }
      ],
      "visualization": "mekkochart"
    },
    "yvyZVN": {
      "id": "yvyZVN",
      "type": "map",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "Distillery"
          ],
          "labels": [
            ""
          ]
        },
        {
          "defs": [
            "Property"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "label": "Value",
          "def": "Sum(Property.Value)"
        }
      ],
      "visualization": "map"
    },
    "zJSbv": {
      "id": "zJSbv",
      "type": "scatterplot",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "AsciiAlpha"
          ],
          "labels": [
            ""
          ]
        }
      ],
      "measures": [
        {
          "def": "Count(AsciiNum)"
        },
        {
          "def": "Count(TransID)"
        }
      ],
      "visualization": "scatterplot"
    }
  },
  "bookmarks": {
    "9d8b2a78-4143-466c-8a5b-ffbf5fd54cba": {
      "id": "9d8b2a78-4143-466c-8a5b-ffbf5fd54cba",
      "title": "BookmarkWithoutSheet",
      "description": "asdfasdfad",
      "selectionFields": "AsciiAlpha"
    },
    "cf5e4644-fd93-4b01-848b-5ba69124987b": {
      "id": "cf5e4644-fd93-4b01-848b-5ba69124987b",
      "title": "BookmarkWithSheet",
      "description": "",
      "sheetId": "f2a50cb3-a7e1-40ac-a015-bc4378773312",
      "selectionFields": "AsciiAlpha"
    }
  }
}`)

func TestConfig_GetSelectables(t *testing.T) {
	var structure appstructure.AppStructure
	if err := json.Unmarshal(structureJSON, &structure); err != nil {
		t.Fatal(err)
	}

	selectables, err := structure.GetSelectables("41dbb01c-d1bd-4528-be05-910ee565988b")
	if err != nil {
		t.Fatal(err)
	}

	expectedSelectables := map[string]interface{}{
		"88a8c69b-5715-4c7c-b4ae-293da7a072d9": nil,
		"fcaf1755-a56b-4d33-a3b4-cebe86254e61": nil,
		"1074afba-0a68-4ef3-b43a-02e500cd2fd8": nil,
		"29ee51e0-f9bb-4a4b-8686-96dcd314727c": nil,
		"532bbe7a-8001-4b94-8b9e-441cc1e8aa48": nil,
		"5932e9a5-14e2-4f6d-8654-c719c288743e": nil,
		"27908a64-d573-4aaa-ac04-b47932ba7995": nil,
		"b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": nil,
		"e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": nil,
	}

	for _, selectable := range selectables {
		_, ok := expectedSelectables[selectable.Id]
		if !ok {
			t.Errorf("object<%s> not expected\n", selectable.Id)
			continue
		}
		delete(expectedSelectables, selectable.Id)
	}

	for id := range expectedSelectables {
		t.Errorf("object<%s> expected but not found\n", id)
	}

	_, err = structure.GetSelectables("not-a-real-object-id")
	switch err.(type) {
	case appstructure.AppStructureObjectNotFoundError:
	default:
		t.Error("Expected AppStructureObjectNotFoundError, got:", err)
	}
}

func TestConfig_GetAppStructures(t *testing.T) {
	cfg, err := NewEmptyConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Scenario = []scenario.Action{
		{
			ActionCore: scenario.ActionCore{
				Type: scenario.ActionOpenApp,
			},
			Settings: scenario.OpenAppSettings{},
		},
		{
			ActionCore: scenario.ActionCore{
				Type: scenario.ActionIterated,
			},
			Settings: scenario.IteratedSettings{
				Iterations: 10,
				Actions: []scenario.Action{
					{
						ActionCore: scenario.ActionCore{
							Type: scenario.ActionClearAll,
						},
						Settings: scenario.ClearAllSettings{},
					},
					{
						ActionCore: scenario.ActionCore{
							Type: scenario.ActionOpenApp,
						},
						Settings: scenario.OpenAppSettings{},
					},
				},
			},
		},
	}

	expectedActions := []string{
		scenario.ActionOpenApp,
		"getappstructure",
		scenario.ActionOpenApp,
		"getappstructure",
	}

	structureScenario := cfg.getAppStructureScenario(false, SummaryTypeNone)

	if len(expectedActions) != len(structureScenario) {
		expected := strings.Join(expectedActions, ",")
		got := make([]string, 0, len(structureScenario))
		for _, act := range structureScenario {
			got = append(got, act.Type)
		}
		t.Fatalf("unexpectedd structure scenario, expected<%s> got<%s>\n", expected, strings.Join(got, ","))
	}

	for i, act := range structureScenario {
		if act.Type != expectedActions[i] {
			t.Errorf("action<%d> expected<%s> got<%s>", i, expectedActions[i], act.Type)
		}
	}
}
