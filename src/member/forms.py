from django import forms
from django.contrib.auth import get_user_model
from django.core.exceptions import ValidationError
from django.utils.translation import gettext_lazy as _


class AddMemberForm(forms.Form):
    username = forms.CharField()
    password = forms.CharField(
        label=_("Password"), strip=False, widget=forms.PasswordInput
    )
    user_model = get_user_model()

    def clean_username(self):
        username = self.cleaned_data.get("username")
        if (
            username
            and self.user_model.objects.filter(username__iexact=username).exists()
        ):
            self.add_error(
                "username", ValidationError(_("The username is not unique"), "username")
            )
        else:
            return username

    def save(self):
        user = self.user_model.objects.create_user(
            self.cleaned_data.get("username"),
            password=self.cleaned_data.get("password"),
        )
        user.save()
        return user
