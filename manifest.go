package main

// Do not change this name, nux need manifest to generate AndroidManifest.xml
const manifest = `
{
    import: {
        ui: "nuxui.org/nuxui/ui",
    },

    application: {
		label: bwakafinder,  

        // application identifier name
        name: "org.nuxui.samples.bwakafinder",
    },

    permissions: [
    ],

    mainWindow: {
        width: 500px,
        height: 700px,
        title: "BW AKA Finder",
        content: {
            type: ui.Layer,
            width: 100%,
            height: 100%,
            children: [
                {
                    type: ui.Column,
                    width: 65%,
                    height: 96%,
                    margin: {left: 1wt, right: 1wt, top: 2px},
                    align: {horizontal: center},
                    children: [
                        {
                            src: "starcaster.png",
                            type: ui.Image,
                            width: 100%,
                            height: 1:1,
                            margin: {top: 4wt, bottom: 3wt}
                        },{
                            type: ui.Text,
                            text: "Add a Browser Source in OBS with the URL as \"localhost:8080\".",
                            font: {size: 15},
                            textColor: #8b8b8b,
                            margin: {bottom: 1wt},
                        }
                    ]
                }

            ]
        }
        background: #000000,
    },
}
`
