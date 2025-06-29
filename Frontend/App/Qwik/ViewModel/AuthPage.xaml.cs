using System.Threading.Tasks;
using Qwik.Model;
using Qwik.Model.Dto;

namespace Qwik;

public partial class AuthPage : ContentPage
{
	public AuthPage()
	{
		InitializeComponent();
	}
    private async void LoginButtonClicked(object sender, EventArgs e)
    {
        try
        {
            AuthenticationReq req = new AuthenticationReq { Login = LoginEnter.Text, Password = PasswordEnter.Text };

            if (string.IsNullOrEmpty(req.Login) || string.IsNullOrEmpty(req.Password))
            {
                await DisplayAlert("Ошибка", "Заполните все поля", "OK");
                return;
            }

            AuthenticationRes body = await Api.SendPostRequest<AuthenticationReq, AuthenticationRes>(req, Api.AuthenticationEndpoint);

            System.Diagnostics.Debug.WriteLine(body);

            Config.JWT = body.Jwt;

            await Shell.Current.Navigation.PopAsync();
        }
        catch (Exception ex) 
        {
            await DisplayAlert("Ошибка", $"{ex.Message}", "OK");
        }
    }

    private async void RegisterButtonClicked(object sender, EventArgs e)
    {
        await Shell.Current.Navigation.PushAsync(new RegPage());
    }
}