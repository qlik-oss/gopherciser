package senseobjdef

var (
	// DefaultListboxDef object definitions for listbox
	DefaultListboxDef = ObjectDef{
		DataDef{DataDefListObject, "/qListObject"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{DataTypeListObject, "/qListObjectDef", DefaultDataHeight},
				},
			},
		},
		&Select{SelectTypeListObjectValues, "/qListObjectDef"},
	}

	// DefaultFilterpane object definitions for
	DefaultFilterpane = ObjectDef{
		DataDef{DataDefNoData, ""},
		nil,
		nil,
	}

	// DefaultBarchart object definitions for barchart
	DefaultBarchart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{DataTypeHyperCubeReducedData, "/qHyperCubeDef", DefaultDataHeight},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultScatterplot object definitions for scatterplot
	DefaultScatterplot = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				&Constraint{
					Path:     "/qHyperCube/qSize/qcy",
					Value:    ">1000",
					Required: true,
				},
				[]GetDataRequests{
					{
						Type: DataTypeHyperCubeBinnedData,
						Path: "/qHyperCubeDef",
					},
				},
			},
			{
				Requests: []GetDataRequests{
					{
						DataTypeHyperCubeData,
						"/qHyperCubeDef",
						1000,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultMap object definitions for map
	DefaultMap = ObjectDef{
		DataDef{DataDefHyperCube, "/qUndoExclude/gaLayers/[0]/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qUndoExclude/gaLayers/0/qHyperCubeDef"},
	}

	// DefaultCombochart object definitions for combochart
	DefaultCombochart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultTable object definitions for table
	DefaultTable = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						DataTypeHyperCubeDataColumns,
						"/qHyperCubeDef",
						40,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeColumnValues, "/qHyperCubeDef"},
	}

	// DefaultPivotTable object definitions for pivot-table
	DefaultPivotTable = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultLinechart object definitions for linechart
	DefaultLinechart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				&Constraint{
					Path:     "/dimensionAxis/continuousAuto",
					Value:    "=true",
					Required: false,
				},
				[]GetDataRequests{
					{
						Type: DataTypeHyperCubeContinuousData,
						Path: "/qHyperCubeDef",
					},
				},
			}, {
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeReducedData,
						Path: "/qHyperCubeDef",
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultPiechart object definitions for piechart
	DefaultPiechart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultTreemap object definitions for treemap
	DefaultTreemap = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultMekkoChart object definitions for mekkochart
	DefaultMekkoChart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qHyperCubeDef"},
	}

	// DefaultTextImage object definitions for text-image
	DefaultTextImage = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	// DefaultKpi object definitions for kpi
	DefaultKpi = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	// DefaultGauge object definitions for gauge
	DefaultGauge = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	// DefaultBoxplot object definitions for boxplot
	DefaultBoxplot = ObjectDef{
		DataDef{DataDefHyperCube, "/qUndoExclude/box/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeData,
						Path: "/qUndoExclude/box/qHyperCubeDef",
					},
					{
						Type: DataTypeHyperCubeStackData,
						Path: "/qUndoExclude/outliers/qHyperCubeDef",
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qUndoExclude/box/qHyperCubeDef"},
	}

	// DefaultDistributionplot object definitions for distributionplot
	DefaultDistributionplot = ObjectDef{
		DataDef{DataDefHyperCube, "/qUndoExclude/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeStackData,
						Path: "/qUndoExclude/qHyperCubeDef",
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qUndoExclude/qHyperCubeDef"},
	}

	// DefaultHistogram object definitions for histogram
	DefaultHistogram = ObjectDef{
		DataDef{DataDefHyperCube, "/qUndoExclude/box/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeData,
						Path: "/qUndoExclude/box/qHyperCubeDef",
					},
				},
			},
		},
		&Select{SelectTypeHypercubeValues, "/qUndoExclude/box/qHyperCubeDef"},
	}

	// DefaultAutoChart object definitions for auto-chart
	DefaultAutoChart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	// DefaultWaterfallChart object definitions for waterfallchart
	DefaultWaterfallChart = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	// DefaultQlikFunnelChartExt object definitions for qlik-funnel-chart-ext
	DefaultQlikFunnelChartExt = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	// DefaultQlikSankeyChartExt object definitions for qlik-sankey-chart-ext
	DefaultQlikSankeyChartExt = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	// Default object definitions for qlik-word-cloud
	DefaultQlikWordCloud = ObjectDef{
		DataDef{DataDefHyperCube, "/qHyperCube"},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	DefaultQlikRadarChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	DefaultQlikBulletChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	DefaultQlikBarplusChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	DefaultQlikMultiKPIChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
	}

	DefaultQlikNetworkChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		&Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	DefaultQlikHeatmapChart = ObjectDef{
		DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		[]Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		nil,
		// heatmap sends two selects, one SelectHyperCubeValues each for dimension 0 and 1 for each selection made.
		// todo support "hardcoded" dimensions and multiple select requests
		//&Select{
		//	Type: SelectTypeHypercubeValues,
		//	Path: "/qHyperCubeDef",
		//},
	}

	DefaultSNOrgChart = ObjectDef{
		DataDef: DataDef{
			Type: DataDefHyperCube,
			Path: "/qHyperCube",
		},
		Data: []Data{
			{
				Requests: []GetDataRequests{
					{
						Type: DataTypeLayout,
					},
				},
			},
		},
		Select: &Select{
			Type: SelectTypeHypercubeValues,
			Path: "/qHyperCubeDef",
		},
	}

	DefaultObjectDefs = ObjectDefs{
		"listbox":               &DefaultListboxDef,
		"filterpane":            &DefaultFilterpane,
		"barchart":              &DefaultBarchart,
		"scatterplot":           &DefaultScatterplot,
		"map":                   &DefaultMap,
		"combochart":            &DefaultCombochart,
		"table":                 &DefaultTable,
		"pivot-table":           &DefaultPivotTable,
		"linechart":             &DefaultLinechart,
		"piechart":              &DefaultPiechart,
		"treemap":               &DefaultTreemap,
		"text-image":            &DefaultTextImage,
		"kpi":                   &DefaultKpi,
		"gauge":                 &DefaultGauge,
		"boxplot":               &DefaultBoxplot,
		"distributionplot":      &DefaultDistributionplot,
		"histogram":             &DefaultHistogram,
		"auto-chart":            &DefaultAutoChart,
		"waterfallchart":        &DefaultWaterfallChart,
		"qlik-funnel-chart-ext": &DefaultQlikFunnelChartExt,
		"qlik-sankey-chart-ext": &DefaultQlikSankeyChartExt,
		"qlik-word-cloud":       &DefaultQlikWordCloud,
		"mekkochart":            &DefaultMekkoChart,
		"qlik-radar-chart":      &DefaultQlikRadarChart,
		"qlik-bullet-chart":     &DefaultQlikBulletChart,
		"qlik-barplus-chart":    &DefaultQlikBarplusChart,
		"qlik-multi-kpi":        &DefaultQlikMultiKPIChart,
		"qlik-network-chart":    &DefaultQlikNetworkChart,
		"qlik-heatmap-chart":    &DefaultQlikHeatmapChart,
		"sn-org-chart":          &DefaultSNOrgChart,
	}
)
