package config

import (
	"github.com/qlik-oss/gopherciser/scenario"
	"strings"
	"testing"
)

var structureJSON = []byte(`{
  "meta": {
    "title": "Ctrl00_allObj",
    "guid": "6c75f3ab-b8e0-46da-912c-8cbf0ecad8f4"
  },
  "objects": {
    "02624185-cc26-4e98-92a1-d2008c36fd85": {
      "id": "02624185-cc26-4e98-92a1-d2008c36fd85",
      "type": "masterobject",
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
    "081e689b-2537-4a68-9091-8e5f6e185c2c": {
      "id": "081e689b-2537-4a68-9091-8e5f6e185c2c",
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
    "1074afba-0a68-4ef3-b43a-02e500cd2fd8": {
      "id": "1074afba-0a68-4ef3-b43a-02e500cd2fd8",
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
      "extendsId": "LczZG"
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
      "extendsId": "WKUFp"
    },
    "2b86d4e7-dff0-4252-9c3f-f67a82626d7d": {
      "id": "2b86d4e7-dff0-4252-9c3f-f67a82626d7d",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Dim2M",
            "description": "Master Dim2",
            "tags": []
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
            "title": "Sum(Exp1)",
            "description": "",
            "tags": []
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
      "extendsId": "bTLJu"
    },
    "53d3f92c-90e5-46a5-ab0f-019e9cf955ec": {
      "id": "53d3f92c-90e5-46a5-ab0f-019e9cf955ec",
      "type": "sheet",
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
      "extendsId": "QquqnR"
    },
    "66c88abf-ee94-4568-8b1b-bff6eeaccf7d": {
      "id": "66c88abf-ee94-4568-8b1b-bff6eeaccf7d",
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
    "6b9169cc-d6de-4fae-9520-d674e2ded148": {
      "id": "6b9169cc-d6de-4fae-9520-d674e2ded148",
      "type": "sheet",
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
      "children": {
        "ynZFW": "mekkochart"
      },
      "selectable": false
    },
    "793a1081-b21a-4dbb-8322-2fb41774452f": {
      "id": "793a1081-b21a-4dbb-8322-2fb41774452f",
      "type": "bookmark",
      "selectable": false
    },
    "7e3f0dd0-b5f2-4179-9172-d48011e167e9": {
      "id": "7e3f0dd0-b5f2-4179-9172-d48011e167e9",
      "type": "filterpane",
      "children": {
        "27908a64-d573-4aaa-ac04-b47932ba7995": "listbox",
        "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": "listbox",
        "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": "listbox"
      },
      "selectable": false,
      "extendsId": "UmDGVm"
    },
    "856be938-e3ae-4437-8752-4af7195cbb78": {
      "id": "856be938-e3ae-4437-8752-4af7195cbb78",
      "type": "sheet",
      "children": {
        "081e689b-2537-4a68-9091-8e5f6e185c2c": "table",
        "66c88abf-ee94-4568-8b1b-bff6eeaccf7d": "pivot-table",
        "9c163098-73e2-480e-8b02-3017742b39be": "pivot-table",
        "b18c7214-33c1-495b-bf17-cbc467389207": "table",
        "b2c3aa99-ba9b-42dc-aa61-c7072157c833": "table",
        "ba3dabae-88a4-41f8-a4d3-a83a3f805232": "pivot-table"
      },
      "selectable": false
    },
    "88a8c69b-5715-4c7c-b4ae-293da7a072d9": {
      "id": "88a8c69b-5715-4c7c-b4ae-293da7a072d9",
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
      "extendsId": "JmdbDg"
    },
    "8b7da3fa-45ee-4920-bd9d-419950be8f2e": {
      "id": "8b7da3fa-45ee-4920-bd9d-419950be8f2e",
      "type": "kpi",
      "selectable": false,
      "visualization": "kpi"
    },
    "9c163098-73e2-480e-8b02-3017742b39be": {
      "id": "9c163098-73e2-480e-8b02-3017742b39be",
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
      "extendsId": "6bd32eab-6ae7-4b80-b3d3-e02f22e1d5b8"
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
            "description": "MasterDim1",
            "tags": []
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
      "extendsId": "ZDxxg"
    },
    "ELvcvsJ": {
      "id": "ELvcvsJ",
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
      "extendsId": "tEwrF"
    },
    "EaUj": {
      "id": "EaUj",
      "type": "map",
      "selectable": false,
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
      "selectable": false,
      "extendsId": "AjLSTb"
    },
    "JMSnZmr": {
      "id": "JMSnZmr",
      "type": "waterfallchart",
      "selectable": false,
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
    "JmdbDg": {
      "id": "JmdbDg",
      "type": "masterobject",
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
    "MJcFdz": {
      "id": "MJcFdz",
      "type": "map",
      "selectable": false,
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
      "extendsId": "bTLJu"
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
      "extendsId": "VqPXn"
    },
    "PaMnJfP": {
      "id": "PaMnJfP",
      "type": "gauge",
      "selectable": false,
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
      "extendsId": "LczZG"
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
      "selectable": false,
      "visualization": "map"
    },
    "QquqnR": {
      "id": "QquqnR",
      "type": "masterobject",
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
      "selectable": false,
      "extendsId": "bMafQ"
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
      "visualization": "text-image"
    },
    "UmDGVm": {
      "id": "UmDGVm",
      "type": "masterobject",
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
      "extendsId": "JmdbDg"
    },
    "VqPXn": {
      "id": "VqPXn",
      "type": "masterobject",
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
            ""
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
      "extendsId": "WKUFp"
    },
    "XwGwH": {
      "id": "XwGwH",
      "type": "auto-chart",
      "selectable": false,
      "visualization": "boxplot"
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
      "extendsId": "ZDxxg"
    },
    "aFjjST": {
      "id": "aFjjST",
      "type": "sheet",
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
    "b18c7214-33c1-495b-bf17-cbc467389207": {
      "id": "b18c7214-33c1-495b-bf17-cbc467389207",
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
    "b2c3aa99-ba9b-42dc-aa61-c7072157c833": {
      "id": "b2c3aa99-ba9b-42dc-aa61-c7072157c833",
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
      "selectable": false,
      "visualization": "gauge"
    },
    "bTLJu": {
      "id": "bTLJu",
      "type": "masterobject",
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
      "extendsId": "QquqnR"
    },
    "ba3dabae-88a4-41f8-a4d3-a83a3f805232": {
      "id": "ba3dabae-88a4-41f8-a4d3-a83a3f805232",
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
    "c10777be-31a4-41f9-a576-797c59820639": {
      "id": "c10777be-31a4-41f9-a576-797c59820639",
      "type": "bookmark",
      "selectable": false
    },
    "c66245be-8472-4900-9f05-e07a0141ed99": {
      "id": "c66245be-8472-4900-9f05-e07a0141ed99",
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
      "extendsId": "VqPXn"
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
    "cwqJNQ": {
      "id": "cwqJNQ",
      "type": "text-image",
      "selectable": false,
      "visualization": "text-image"
    },
    "d582f067-d8a0-47b4-b62f-7930ce009da7": {
      "id": "d582f067-d8a0-47b4-b62f-7930ce009da7",
      "type": "bookmark",
      "selectable": false
    },
    "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33": {
      "id": "d86adb80-fdb0-4da5-9d63-d95ba3f2bd33",
      "type": "dimension",
      "selectable": false,
      "dimensions": [
        {
          "meta": {
            "title": "Dims123D",
            "description": "DrillDown Dims",
            "tags": []
          },
          "defs": [
            "Dim1",
            "Dim2",
            "Dim3"
          ]
        }
      ]
    },
    "d8ac3450-c845-45fb-95d7-711bfaef31b9": {
      "id": "d8ac3450-c845-45fb-95d7-711bfaef31b9",
      "type": "bookmark",
      "selectable": false
    },
    "dYZkN": {
      "id": "dYZkN",
      "type": "sheet",
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
      "selectable": false,
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
    "e94a7e71-0ebb-47c0-aeea-bd4b9eb8c69d": {
      "id": "e94a7e71-0ebb-47c0-aeea-bd4b9eb8c69d",
      "type": "text-image",
      "selectable": false,
      "extendsId": "AjLSTb"
    },
    "e98010b2-a77e-4c5c-9612-1c3329713141": {
      "id": "e98010b2-a77e-4c5c-9612-1c3329713141",
      "type": "sheet",
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
      "selectable": false,
      "extendsId": "dmjZp"
    },
    "fa4229b9-6e9b-4bea-b516-c9c717167a29": {
      "id": "fa4229b9-6e9b-4bea-b516-c9c717167a29",
      "type": "gauge",
      "selectable": false,
      "extendsId": "bMafQ"
    },
    "fcaf1755-a56b-4d33-a3b4-cebe86254e61": {
      "id": "fcaf1755-a56b-4d33-a3b4-cebe86254e61",
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
      "extendsId": "tEwrF"
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
      "visualization": "text-image"
    },
    "gJfAjj": {
      "id": "gJfAjj",
      "type": "sheet",
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
            "title": "Avg(Exp2)",
            "description": "",
            "tags": []
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
      "visualization": "text-image"
    },
    "hQDEATt": {
      "id": "hQDEATt",
      "type": "sheet",
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
            "=Class([Num],5)"
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
      "children": {
        "27908a64-d573-4aaa-ac04-b47932ba7995": "listbox",
        "b3e6f4a3-b43c-49d8-ae5f-87586a937b3a": "listbox",
        "e13b12f0-c2c1-42cd-aa54-ef6efdbf24c0": "listbox"
      },
      "selectable": false,
      "extendsId": "UmDGVm"
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
      "selectable": false,
      "visualization": "map"
    },
    "pHjJnUa": {
      "id": "pHjJnUa",
      "type": "histogram",
      "selectable": true,
      "dimensions": [
        {
          "defs": [
            "=Class([Expression2],0.65)"
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
            "title": "Sum(Exp1)",
            "description": "",
            "tags": []
          },
          "label": "Sum(Exp1)",
          "def": "Sum(Expression1)"
        }
      ]
    },
    "ppUMK": {
      "id": "ppUMK",
      "type": "kpi",
      "selectable": false,
      "extendsId": "dmjZp"
    },
    "ppjDSJ": {
      "id": "ppjDSJ",
      "type": "sheet",
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
      "selectable": false,
      "visualization": "map"
    },
    "rkuwZh": {
      "id": "rkuwZh",
      "type": "sheet",
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
      "extendsId": "35585276-91fd-45f2-8594-e28feae7a8cd"
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
      "selectable": false,
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
            "title": "Max(Exp3)",
            "description": "",
            "tags": []
          },
          "label": "Max(Exp3)",
          "def": "Max(Expression3)"
        }
      ]
    },
    "uaHYE": {
      "id": "uaHYE",
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
      "extendsId": "02624185-cc26-4e98-92a1-d2008c36fd85"
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
            "=Class([Expression1],60)"
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
      "selectable": false,
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
    "793a1081-b21a-4dbb-8322-2fb41774452f": {
      "id": "793a1081-b21a-4dbb-8322-2fb41774452f",
      "title": "PrivateBookmarkNoSheet",
      "description": "",
      "selectionFields": "TransID"
    },
    "c10777be-31a4-41f9-a576-797c59820639": {
      "id": "c10777be-31a4-41f9-a576-797c59820639",
      "title": "PublicBookmarkNoSheet",
      "description": "",
      "selectionFields": "TransID"
    },
    "d582f067-d8a0-47b4-b62f-7930ce009da7": {
      "id": "d582f067-d8a0-47b4-b62f-7930ce009da7",
      "title": "PublicBookmarkWithSheet",
      "description": "",
      "sheetId": "hQDEATt",
      "selectionFields": "TransID"
    },
    "d8ac3450-c845-45fb-95d7-711bfaef31b9": {
      "id": "d8ac3450-c845-45fb-95d7-711bfaef31b9",
      "title": "PrivateBookmarkWithSheet",
      "description": "",
      "sheetId": "hQDEATt",
      "selectionFields": "TransID"
    }
  }
}`)

func TestConfig_GetSelectables(t *testing.T) {
	var structure AppStructure
	if err := jsonit.Unmarshal(structureJSON, &structure); err != nil {
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

	structureScenario := cfg.getAppStructureScenario()

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
