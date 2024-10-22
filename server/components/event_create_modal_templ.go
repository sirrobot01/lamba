// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func CreateEventModal() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"fixed inset-0 bg-black bg-opacity-50 hidden items-center justify-center\" id=\"createEventModal\"><div class=\"bg-white p-6 rounded-lg w-[600px]\"><div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-xl font-semibold\">Create Event</h2><button class=\"text-gray-500 hover:text-gray-700\" onclick=\"closeEventModal();\">✕</button></div><form hx-post=\"/events\" hx-target=\"#main-content\" hx-swap=\"outerHTML\" hx-on::after-request=\"closeEventModal()\"><div class=\"space-y-4\"><div><label class=\"block text-sm font-medium text-gray-700\">Event Name</label> <input type=\"text\" name=\"name\" required class=\"mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-4 py-2\"></div><div><label class=\"block text-sm font-medium text-gray-700\">Payload</label> <textarea name=\"payload\" class=\"mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500\"></textarea></div></div><div class=\"mt-6 flex justify-end space-x-3\"><button type=\"button\" onclick=\"closeEventModal()\" class=\"px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50\">Cancel</button> <button type=\"submit\" class=\"px-4 py-2 text-sm font-medium text-white bg-orange-500 rounded-md hover:bg-orange-600\">Create</button></div></form></div></div><script>\n        function closeEventModal() {\n            document.getElementById('createEventModal').style.display = 'none';\n        }\n\n        function showEventModal() {\n            document.getElementById('createEventModal').style.display = 'flex';\n        }\n    </script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
