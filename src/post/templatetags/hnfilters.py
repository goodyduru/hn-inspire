from django import template


register = template.Library()

@register.filter(name="indent")
def indent(value):
    if value > 1:
        return f"margin-left: {value-1}em"
    return ""
