package bpmetadata

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/util"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

//go:embed schema/gcp-blueprint-metadata.json
var s []byte

// validateMetadata validates the metadata files for the provided
// blueprint path. This validation occurs for top-level blueprint
// metadata and blueprints in the modules/ folder, if present
func validateMetadata(bpPath, wdPath string) error {
	// load schema from the binary
	schemaLoader := gojsonschema.NewStringLoader(string(s))

	// check if the provided output path is relative
	if !path.IsAbs(bpPath) {
		bpPath = path.Join(wdPath, bpPath)
	}

	// We don't need to validate metadata under .terraform folders
	skipDirsToValidate := []string{".terraform/"}
	metadataFiles, err := util.FindFilesWithPattern(bpPath, `^metadata(?:.display)?.yaml$`, skipDirsToValidate)
	if err != nil {
		Log.Error("unable to read at: %s", bpPath, "err", err)
	}

	var vErrs []error
	for _, f := range metadataFiles {
		err = validateMetadataYaml(f, schemaLoader)
		if err != nil {
			vErrs = append(vErrs, err)
			Log.Error("core metadata validation failed", "err", err)
		}
	}

	if len(vErrs) > 0 {
		return fmt.Errorf("metadata validation failed for at least one blueprint")
	}

	return nil
}

// validateMetadata validates an individual yaml file present at path "m"
func validateMetadataYaml(m string, schema gojsonschema.JSONLoader) error {
	// prepare metadata for validation by converting it from YAML to JSON
	mBytes, err := convertYamlToJson(m)
	if err != nil {
		return fmt.Errorf("yaml to json conversion failed for metadata at path %s. error: %w", m, err)
	}

	// load metadata from the path
	yamlLoader := gojsonschema.NewStringLoader(string(mBytes))

	// validate metadata against the schema
	result, err := gojsonschema.Validate(schema, yamlLoader)
	if err != nil {
		return fmt.Errorf("metadata validation failed for %s. error: %w", m, err)
	}

	if !result.Valid() {
		for _, e := range result.Errors() {
			Log.Error("validation error", "err", e)
		}

		return fmt.Errorf("metdata validation failed for: %s", m)
	}

	Log.Info("metadata is valid", "path", m)
	return nil
}

// prepares metadata bytes for validation since direct
// validation of YAML is not possible
func convertYamlToJson(m string) ([]byte, error) {
	// read metadata for validation
	b, err := os.ReadFile(m)
	if err != nil {
		return nil, fmt.Errorf("unable to read metadata at path %s. error: %w", m, err)
	}

	if len(b) == 0 {
		return nil, fmt.Errorf("metadata contents can not be empty")
	}

	json, err := yaml.YAMLToJSON(b)
	if err != nil {
		return nil, fmt.Errorf("metadata contents are invalid: %s", err.Error())
	}

	return json, nil
}

func validataMetadataForADC(rootModulePath string)error{

	// load schema from the binary
	schemaLoader := gojsonschema.NewStringLoader(string(s))

  metadataFiles:=[]string{
		path.Join(rootModulePath, "metadata.yaml"),
	}

	var vErrs []error
	for _, f := range metadataFiles {
		// validate schema
		err := validateMetadataYaml(f, schemaLoader)
		if err != nil {
			vErrs = append(vErrs, err)
			Log.Error("core metadata validation failed", "err", err)
		}
	}
 // validate fields from metadata.yaml
 // validate connection, roles, apis, output type, defaults

//  bpObj, err :=  UnmarshalMetadata(rootModulePath, "metadata.yaml")
//  if err != nil && !errors.Is(err, os.ErrNotExist) {
// 	 return err
//  }


//  outputHavingMissingTypes:= [] *BlueprintOutput
//  for _, output :=range bpObj.Spec.Interfaces.Outputs{
//        if (output.Type==nil){
// 				outputHavingMissingTypes=append(outputHavingMissingTypes, output)
// 			 }
//  }
//  Log.Info("outputHavingMissingTypes content: "+string(outputHavingMissingTypes))

 return nil
}


func ValidateRootModuleForADC(bpPath string) error{

	// files root module must have TODO: Add description describing need of each file
	requiredFilesForRoot:=[]string {"README.md", "main.tf", "versions.tf"}
	missingFiles:=checkFilePresence(bpPath, requiredFilesForRoot)

	if len(missingFiles) > 0 {
		return fmt.Errorf("top-level module must have following files also:%s", missingFiles)
	}

	goodToHaveFiles:=[]string{"test/setup/iam.tf",  "test/setup/main.tf"}
	missingGoodToHaveFiles:=checkFilePresence(bpPath, goodToHaveFiles)

	if len(missingGoodToHaveFiles) > 0 {
		Log.Warn("It is good to have these files also for generating metadata: ["+ strings.Join(missingGoodToHaveFiles, ", ") +"]\n")
	}

	otherFiles:=[]string{"assets/icon.png", "modules/", "examples",}
	missingOtherFiles:=checkFilePresence(bpPath, otherFiles)
	if len(missingOtherFiles) > 0 {
		Log.Info("Cft module can have these files also: [" + strings.Join(missingOtherFiles, ", ")+"]\n")
	}

 // Validate Readme.md file content
 err:= validateReadme(path.Join(bpPath, readmeFileName))
 if err !=nil{
	 return err
 }

 // its not necessary that user is naming his file as per our convention e.g.
 // it could be vars.tf or out.tf
 // err = validateVariablesFiles(path.Join(bpPath, "variables.tf"))
 // if err !=nil{
 // 	return err
 // }
 // err= validateOutputFiles(path.Join(bpPath, "outputs.tf"))
 // if err !=nil{
 // 	return err
 // }

 err = validateVersionsFiles(path.Join(bpPath, "versions.tf"))
 if err !=nil {
	 return err
 }

 return nil
}


// func validateVariablesFiles(variablesFilePath string ) error{
// 	Log.Info("**************** Validating  variables.tf****************")
// 	//Default variables thing
// 	// valdate the schema

// 	Log.Info("****************Done Validating  variables.tf****************\n\n")
// return nil
// }

// func validateOutputFiles(outputFilePath string ) error{
// 	Log.Info("**************** Validating  outputs.tf****************")
// 	// validate the schema and create the output type of the output but output type
// 	//is generated from state file of test

// 	Log.Info("****************Done Validating  outputs.tf****************\n\n")
// return nil
// }

func validateVersionsFiles(versionsConfigPath string)error {
 Log.Info("**************** Validating  versions.tf****************")

 //set of allowed providers
 allowedTerraformProviders:=	[]string{
	 "hashicorp/google","hashicorp/google-beta",
 }

 p := hclparse.NewParser()
 versionsFile, diags := p.ParseHCLFile(versionsConfigPath)
 err := hasHclErrors(diags)
 if err != nil {
	 return  err
 }

 // parse out the required providers from the config
 var hclModule tfconfig.Module
 hclModule.RequiredProviders = make(map[string]*tfconfig.ProviderRequirement)
 diags = tfconfig.LoadModuleFromFile(versionsFile, &hclModule)
 err = hasHclErrors(diags)
 if err != nil {
	 return  err
 }

	var errors []string
 for _, providerData := range hclModule.RequiredProviders {
	 if providerData.Source == "" {
		 errors=append(errors, fmt.Sprintln("source not found in provider settings"))
	 }else if slices.Index(allowedTerraformProviders, providerData.Source)==-1{
		 errors=append(errors, fmt.Sprintln("Incorrect terraform provider: "+providerData.Source + ". Only following terraform providers are allowed as of now: ["+ strings.Join(allowedTerraformProviders, ", ")+"]"))
	 }

	 if len(providerData.VersionConstraints) == 0 {
		 errors=append(errors, fmt.Sprintln("version not found in provider settings"))
	 }
 }

if len(errors)>0{
 return fmt.Errorf(strings.Join(errors, "\n"))
}

Log.Info("****************Done Validating  versions.tf****************\n\n")
return nil
}


func validateReadme(readmeFilePath string ) error{
 Log.Info("\n\n**************** Validating  README.md****************")
 readmeContent, err := os.ReadFile(readmeFilePath)
 if err != nil {
	 return  fmt.Errorf("blueprint readme markdown is missing, create one using https://tinyurl.com/tf-mod-readme | error: %w", err)
 }
 /// Repo details validation

 // Must have headings
	mustHaveHeadings:= []markdownHeading{
	 {
		 headLevel:1,
		 headOrder:1,
		 headTitle:"",
		 content:false,
		 description:"Title of the module",

	 },
	}

 missingMustHaveHeadings:=missingHeadings(readmeContent,mustHaveHeadings)
 var errors []string
 if len(missingMustHaveHeadings)>0{
	 for _, heading :=range missingMustHaveHeadings{
		 e:= fmt.Sprintf("%s\t %s",heading.headTitle, heading.description)
		 errors = append(errors, e)
	 }
	 if len(errors) > 0 {
		 return fmt.Errorf("%s", strings.Join(errors, "\n"))
	 }
 }

 goodToHaveMarkdownHeadings:= []markdownHeading{
	 {
		 headLevel:1,
		 headOrder:1,
		 headTitle:"",
		 content:false,
		 description:"Title of the module",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Tagline",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.description.Tagline` field in metadata.yaml",

	 },
 }
 missingGoodToHaveHeadings:=missingHeadings( readmeContent,goodToHaveMarkdownHeadings)


 if len(missingGoodToHaveHeadings)>0{
	 var warningMsg []string
	 for _, heading :=range missingGoodToHaveHeadings{
		 warningMsg=append(warningMsg,"# "+heading.headTitle+": \t"+ heading.description)
		 // Log.Warn("# "+heading.headTitle+": \t"+ heading.description)/
	 }
	 Log.Warn("\nGood to have headings- \n"+ strings.Join(warningMsg, "\n")+"\n")
 }

	otherHeadings:= []markdownHeading{
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Tagline",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.description.Tagline` field in metadata.yaml",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Detailed",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.Description.Detailed` field in metadata.yaml",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"PreDeploy",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.Description.PreDeploy` field in metadata.yaml",

	 },

	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Architecture",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.Description.Architecture` field in metadata.yaml",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Deployment Duration",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.DeploymentDuration` field in metadata.yaml",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Cost",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintInfo.CostEstimate` field in metadata.yaml",

	 },
	 {
		 headLevel:-1,
		 headOrder:-1,
		 headTitle:"Documentation",
		 content:true,
		 description:"corresponds to `BlueprintMetadataSpec.BlueprintContent.Documentation` field in metadata.yaml,"+
									 "if Documentation heading is present as a root chidren of markdown the very next paragraph is scanned,"+
									 "expecting to find an image or a link (as the diagram) followed immediately by a block of text"+
									 "(as the description). If this structure is present, it extracts the diagram URL and"+
									 "the description lines; otherwise, it does not include this field in the metadata.yaml.",

	 },
 }
 otherMissingHeadings:=missingHeadings( readmeContent,otherHeadings)

 if len(otherMissingHeadings)>0 {
	 var infoMsg []string
	 for _, heading :=range otherMissingHeadings{
		 infoMsg=append(infoMsg,"# "+heading.headTitle+ ":\t"+ heading.description )
	 }
	 Log.Info("\nOther missing headings-\n"+ strings.Join(infoMsg, "\n")+"\n")
 }
 Log.Info("\n****************Done Validating  README.md****************\n\n")
return nil
}

func missingHeadings(content []byte, markdownHeadings []markdownHeading)[]markdownHeading{
 missingHeadings:=[]markdownHeading{}
 for _, heading:= range markdownHeadings{
	 _, err := getMdContent(content, heading.headLevel, heading.headOrder, heading.headTitle, heading.content)
	 if err != nil {
		 missingHeadings=append(missingHeadings, heading)
	 }
 }
 return missingHeadings
}

func checkFilePresence(bpPath string, filePaths[] string) []string{

 missingFiles:=[]string{}
 for _, fileName := range filePaths{
	 filePath := path.Join(bpPath,fileName)
	 _, err := os.Stat(filePath)

	 if err != nil {
		 missingFiles = append(missingFiles, fileName)
	 }

 }
 return missingFiles
}

