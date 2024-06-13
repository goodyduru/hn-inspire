from django.contrib import admin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.contrib.auth.models import User

from .models import Profile


# Register your models here.
class MemberInline(admin.StackedInline):
    model = Profile
    can_delete = False
    verbose_name_plural = "members"


class UserAdmin(BaseUserAdmin):
    inlines = [MemberInline]


admin.site.unregister(User)
admin.site.register(User, UserAdmin)
